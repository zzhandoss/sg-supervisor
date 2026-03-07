package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"sg-supervisor/internal/app"
)

func runBuildRelease(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("build-release", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	platform := fs.String("platform", "", "target platform: windows or linux")
	version := fs.String("version", "", "release version")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *platform == "" {
		return errors.New("platform is required")
	}
	if *version == "" {
		return errors.New("version is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.BuildRelease(ctx, *platform, *version, *binaryPath)
	if err != nil {
		return err
	}
	fmt.Println(report.ArtifactPath)
	for _, warning := range report.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	return nil
}

func runBuildReleaseSet(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("build-release-set", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	version := fs.String("version", "", "release version")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *version == "" {
		return errors.New("version is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.BuildReleaseSet(ctx, *version, *binaryPath)
	if err != nil {
		return err
	}
	fmt.Println(report.MetadataPath)
	for _, warning := range report.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	return nil
}
