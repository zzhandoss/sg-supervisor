package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"sg-supervisor/internal/app"
)

func runInstallPackage(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("install-package", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	packageID := fs.String("package-id", "", "imported package id")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *packageID == "" {
		return errors.New("package-id is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.InstallPackage(ctx, *packageID, *binaryPath)
	if err != nil {
		return err
	}
	fmt.Println(report.ServiceName)
	return nil
}

func runRepair(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("repair", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	binaryPath := fs.String("binary", "", "supervisor binary path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.Repair(ctx, *binaryPath)
	if err != nil {
		return err
	}
	fmt.Println(report.ActivePackageID)
	return nil
}

func runUninstall(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	mode := fs.String("mode", "keep-state", "uninstall mode")
	if err := fs.Parse(args); err != nil {
		return err
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	report, err := supervisor.Uninstall(ctx, *mode)
	if err != nil {
		if len(report.Issues) > 0 || len(report.RemovedPaths) > 0 || report.Completed {
			fmt.Fprintln(os.Stderr, report.Mode)
			for _, issue := range report.Issues {
				fmt.Fprintf(os.Stderr, "%s: %s\n", issue.Step, issue.Message)
			}
		}
		return err
	}
	fmt.Println(report.Mode)
	return nil
}
