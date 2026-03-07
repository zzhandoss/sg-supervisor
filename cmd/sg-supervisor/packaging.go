package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"sg-supervisor/internal/app"
)

func runAssemblePackage(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("assemble-package", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	platform := fs.String("platform", "", "target platform: windows or linux")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *platform == "" {
		return errors.New("platform is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.AssemblePackage(ctx, *platform, *binaryPath)
	if err != nil {
		return err
	}
	fmt.Println(report.OutputDir)
	return nil
}
