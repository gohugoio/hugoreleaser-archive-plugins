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
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	// Hugoreleaser API

	"github.com/bep/execrpc"
	"github.com/gohugoio/hugoreleaser-plugins-api/archiveplugin"
	"github.com/gohugoio/hugoreleaser-plugins-api/model"

	// nfpm
	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/deb" // init format
	"github.com/goreleaser/nfpm/v2/files"
)

const name = "deb"

func main() {
	var archiveClient archiveClient
	server, err := execrpc.NewServer(
		execrpc.ServerOptions[model.Config, archiveplugin.Request, any, model.Receipt]{
			GetHasher:     nil,
			DelayDelivery: false,
			Init: func(v model.Config) error {
				archiveClient.cfg = v
				return nil
			},
			Handle: func(call *execrpc.Call[archiveplugin.Request, any, model.Receipt]) {
				model.Infof(call, "Creating archive %s", call.Request.OutFilename)
				var receipt model.Receipt
				if !archiveClient.cfg.Try {
					if err := archiveClient.createArchive(call.Request); err != nil {
						receipt.Error = model.NewError(name, err)
					}
				}
				receipt = <-call.Receipt()
				call.Close(false, receipt)
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

// Settings is fetched from archive_settings.custom_settings in the archive configuration.
type Settings struct {
	Vendor      string
	Homepage    string
	Maintainer  string
	Description string
	License     string

	PackageName     string
	Section         string
	Priority        string
	Epoch           string
	Release         string
	Prerelease      string
	VersionMetadata string
}

type archiveClient struct {
	cfg model.Config
}

type debArchivist struct {
	out         io.WriteCloser
	files       files.Contents
	projectInfo model.ProjectInfo
	goInfo      model.GoInfo
	settings    Settings
}

func (a *debArchivist) Add(sourceFilename, targetPath string, mode fs.FileMode) error {
	a.files = append(a.files, &files.Content{
		Source:      filepath.ToSlash(sourceFilename),
		Destination: targetPath,
		FileInfo: &files.ContentFileInfo{
			Mode: mode,
		},
	})

	return nil
}

func (a *debArchivist) Finalize() error {
	s := a.settings
	g := a.goInfo
	p := a.projectInfo

	if s.PackageName == "" {
		s.PackageName = p.Project
	}

	info := &nfpm.Info{
		Platform:        g.Goos,
		Arch:            g.Goarch,
		Name:            s.PackageName,
		Version:         p.Tag,
		Section:         s.Section,
		Priority:        s.Priority,
		Epoch:           s.Epoch,
		Release:         s.Release,
		Prerelease:      s.Prerelease,
		VersionMetadata: s.VersionMetadata,
		Maintainer:      s.Maintainer,
		Description:     s.Description,
		Vendor:          s.Vendor,
		Homepage:        s.Homepage,
		License:         s.License,
		Overridables: nfpm.Overridables{
			Contents: a.files,
		},
	}

	packager, err := nfpm.Get("deb")
	if err != nil {
		return err
	}

	info = nfpm.WithDefaults(info)

	return packager.Package(info, a.out)
}

func (c archiveClient) createArchive(req archiveplugin.Request) error {
	if err := req.Init(); err != nil {
		return err
	}

	f, err := os.Create(req.OutFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	settings, err := model.FromMap[any, Settings](req.Settings)
	if err != nil {
		return err
	}

	archivist := &debArchivist{
		out:         f,
		projectInfo: c.cfg.ProjectInfo,
		goInfo:      req.GoInfo,
		settings:    settings,
	}

	for _, file := range req.Files {
		if file.Mode == 0 {
			file.Mode = 0o644
		}
		archivist.Add(file.SourcePathAbs, file.TargetPath, file.Mode)
	}

	return archivist.Finalize()
}
