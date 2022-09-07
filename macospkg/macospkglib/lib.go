package macospkglib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bep/buildpkg"
)

// BuildPkg creates a MacOS package.
func BuildPkg(
	settings Settings,
	infof func(format string, args ...interface{}),
	extractToDir func(dir string) error,
	outFilename string) (string, error) {

	if err := settings.Init(); err != nil {
		return "", err
	}

	workingDir, err := os.MkdirTemp("", "macospkglib")
	if err != nil {
		return "", err
	}

	// All the files in here will be packaged.
	stagingDir := filepath.Join(workingDir, "staging")
	if err := os.MkdirAll(stagingDir, 0777); err != nil {
		return "", err
	}

	if err := extractToDir(stagingDir); err != nil {
		return "", err
	}

	if outFilename == "" {
		outFilename = filepath.Join(workingDir, "mypack.pkg")

	}

	if infof == nil {
		infof = log.Printf
	}

	opts := buildpkg.Options{
		Infof:                 infof,
		Identifier:            settings.PackageIdentifier,
		Version:               settings.PackageVersion,
		InstallLocation:       "/usr/local/bin",
		Dir:                   workingDir,
		SigningIdentity:       settings.AppleSigningIdentity,
		StagingDirectory:      stagingDir,
		PackageOutputFilename: outFilename,
	}

	builder, err := buildpkg.New(opts)
	if err != nil {
		return "", err
	}

	if err = builder.Build(); err != nil {
		return "", err
	}

	return outFilename, err

}

// Settings is fetched from archive_settings.custom_settings in the archive configuration.
type Settings struct {
	// Package settings. We pass this on to the buildpkg tool.
	PackageSettings

	// Only used for the local plugin.
	NeedsAppleSettings   bool
	AppleSigningIdentity string `mapstructure:"apple_signing_identity"`

	// AWS settings.
	NeedsAWSSettings bool
	Bucket           string `mapstructure:"bucket"`
	Queue            string `mapstructure:"queue"`
	AccessKeyID      string `mapstructure:"access_key_id"`
	SecretAccessKey  string `mapstructure:"secret_access_key"`
}

func (s *Settings) Init() error {
	what := "archive_settings.custom_settings"
	if s.NeedsAppleSettings {
		if s.AppleSigningIdentity == "" {
			return fmt.Errorf("%s.apple_signing_identity is required", what)
		}
	}
	if s.NeedsAWSSettings {
		if s.Bucket == "" {
			return fmt.Errorf("%s.bucket is required", what)
		}
		if s.Queue == "" {
			return fmt.Errorf("%s.queue is required", what)
		}
		if s.AccessKeyID == "" {
			return fmt.Errorf("%s.access_key_id is required", what)
		}
		if s.SecretAccessKey == "" {
			return fmt.Errorf("%s.secret_access_key is required", what)
		}
	}

	if s.PackageSettings.PackageIdentifier == "" {
		return fmt.Errorf("%s.package_identifier is required", what)
	}
	if s.PackageSettings.PackageVersion == "" {
		return fmt.Errorf("%s.package_version is required", what)
	}
	return nil

}

type PackageSettings struct {
	// Package settings.
	// E.g. io.gohugo.hugoreleaser
	PackageIdentifier string `mapstructure:"package_identifier"`
	// E.g. v0.1.0
	PackageVersion string `mapstructure:"package_version"`
}
