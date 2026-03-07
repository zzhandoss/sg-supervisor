package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"sg-supervisor/internal/app"
)

func runRenderServiceHost(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("render-service-host", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	artifacts, err := supervisor.RenderServiceHostArtifacts(ctx, *binaryPath)
	if err != nil {
		return err
	}
	if len(artifacts.WrittenFiles) == 0 {
		return errors.New("service host artifacts were not written")
	}
	fmt.Println(artifacts.LinuxUnitPath)
	return nil
}
