// Copyright 2022 The Hugoreleaser Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gohugoio/hugoreleaser-archive-plugins/macospkg/macospkglib"

	"github.com/bep/helpers/archivehelpers"
	"github.com/bep/s3rpc"
	"golang.org/x/sync/errgroup"

	// Hugoreleaser API

	"github.com/gohugoio/hugoreleaser-plugins-api/archiveplugin"
	"github.com/gohugoio/hugoreleaser-plugins-api/model"
	"github.com/gohugoio/hugoreleaser-plugins-api/server"
)

const (
	pluginName = "macospkgremote"
)

func main() {

	if len(os.Args) > 1 && os.Args[1] == "localserver" {
		if err := startLocalServer(); err != nil {
			log.Fatal(err)
		}
		return
	}

	server, err := server.New(
		func(d server.Dispatcher, req archiveplugin.Request) archiveplugin.Response {
			d.Infof("Creating archive %s", req.OutFilename)

			if err := req.Init(); err != nil {
				return errResponse(err)
			}

			if len(req.Files) != 1 {
				return errResponse(fmt.Errorf("this plugin currently support 1 file only (the binary), got %d", len(req.Files)))
			}

			if err := createArchive(d, req); err != nil {
				return errResponse(err)
			}

			// Empty response is a success.
			return archiveplugin.Response{}
		},
	)
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}

	_ = server.Wait()
}

func createArchive(d server.Dispatcher, req archiveplugin.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	settings, err := model.FromMap[any, macospkglib.Settings](req.Settings)
	if err != nil {
		return err
	}
	settings.NeedsAWSSettings = true
	if err := settings.Init(); err != nil {
		return err
	}

	client, err := s3rpc.NewClient(
		s3rpc.ClientOptions{
			Queue:   settings.Queue,
			Timeout: 5 * time.Minute,
			Infof:   d.Infof,
			AWSConfig: s3rpc.AWSConfig{
				Bucket:          settings.Bucket,
				AccessKeyID:     settings.AccessKeyID,
				SecretAccessKey: settings.SecretAccessKey,
			},
		},
	)

	if err != nil {
		return err
	}

	// Put the file in a gzipped tarball to preserve file permissions and make it faster to upload.
	archiver, err := archivehelpers.New(archivehelpers.TypeTarGz)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "hugoreleaser-macospkgremote.*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	file := req.Files[0]
	// TODO(bep) file modes.
	if err := archiver.ArchiveDirectory(filepath.Dir(file.SourcePathAbs), func(s string) bool { return s == file.SourcePathAbs }, tempFile); err != nil {
		return err
	}

	metadata := map[string]string{
		"package_identifier": settings.PackageIdentifier,
		"package_version":    settings.PackageVersion,
	}

	res, err := client.Execute(ctx, pluginName, s3rpc.Input{Filename: tempFile.Name(), Metadata: metadata})
	if err != nil {
		return err
	}

	resf, err := os.Open(res.Filename)
	if err != nil {
		return err
	}
	defer resf.Close()

	outf, err := os.Create(req.OutFilename)
	if err != nil {
		return err
	}
	defer outf.Close()
	_, err = io.Copy(outf, resf)

	return err
}

func startLocalServer() error {
	signingID := os.Getenv("BUILDPKG_APPLE_DEVELOPER_SIGNING_IDENTITY")
	if signingID == "" {
		return errors.New("BUILDPKG_APPLE_DEVELOPER_SIGNING_IDENTITY not set in environment. Must be set to a valid  Developer ID Application + Developer ID Installer signing identity.")
	}
	infol := log.Printf

	handlers := s3rpc.Handlers{
		pluginName: func(ctx context.Context, input s3rpc.Input) (s3rpc.Output, error) {

			infol("%s", pluginName)

			settings, err := model.FromMap[string, macospkglib.Settings](input.Metadata)
			if err != nil {
				return s3rpc.Output{}, err
			}

			archiver, err := archivehelpers.New(archivehelpers.TypeTarGz)
			if err != nil {
				return s3rpc.Output{}, err
			}

			r, err := os.Open(input.Filename)
			if err != nil {
				return s3rpc.Output{}, err
			}

			filename, err := macospkglib.BuildPkg(
				settings,
				nil,
				func(dir string) error {
					return archiver.Extract(r, dir)
				},
				"",
			)

			if err != nil {
				return s3rpc.Output{}, err
			}

			return s3rpc.Output{
				Filename: filename,
				Metadata: nil,
			}, nil

		},
	}

	server, err := s3rpc.NewServer(
		s3rpc.ServerOptions{
			Handlers:     handlers,
			Queue:        os.Getenv("S3RPC_SERVER_QUEUE"),
			Infof:        infol,
			PollInterval: 45 * time.Second,
			AWSConfig: s3rpc.AWSConfig{
				Bucket:          "s3fptest",
				AccessKeyID:     os.Getenv("S3RPC_SERVER_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("S3RPC_SERVER_SECRET_ACCESS_KEY"),
			},
		},
	)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errWg, errCtx := errgroup.WithContext(ctx)
	errWg.Go(func() error {
		return server.ListenAndServe(errCtx)
	})

	errWg.Go(func() error {
		<-errCtx.Done()
		stop()
		infol("Closing server ...")
		return server.Close()
	})

	err = errWg.Wait()

	if err != nil && !errors.Is(err, context.Canceled) {
		return err

	}

	infol("Done.")

	return nil

}

func errResponse(err error) archiveplugin.Response {
	return archiveplugin.Response{Error: model.NewError(pluginName, err)}
}
