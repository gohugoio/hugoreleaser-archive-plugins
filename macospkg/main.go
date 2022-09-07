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
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gohugoio/hugoreleaser-archive-plugins/macospkg/macospkglib"

	// Hugoreleaser API

	"github.com/gohugoio/hugoreleaser-plugins-api/archiveplugin"
	"github.com/gohugoio/hugoreleaser-plugins-api/model"
	"github.com/gohugoio/hugoreleaser-plugins-api/server"
)

const (
	pluginName = "macospkg"
)

func main() {

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
	settings, err := model.FromMap[any, macospkglib.Settings](req.Settings)
	if err != nil {
		return err
	}
	settings.NeedsAppleSettings = true
	_, err = macospkglib.BuildPkg(
		settings,
		d.Infof,
		func(dir string) error {
			file := req.Files[0]
			target, err := os.OpenFile(filepath.Join(dir, file.TargetPath), os.O_CREATE|os.O_WRONLY, file.Mode)
			if err != nil {
				return err
			}
			defer target.Close()
			sourc, err := os.Open(file.SourcePathAbs)
			if err != nil {
				return err
			}
			defer sourc.Close()
			_, err = io.Copy(target, sourc)
			return err

		},
		req.OutFilename,
	)

	return err
}

func errResponse(err error) archiveplugin.Response {
	return archiveplugin.Response{Error: model.NewError(pluginName, err)}
}
