package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"sg-supervisor/internal/app"
)

func runSetSetupField(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("set-setup-field", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	key := fs.String("key", "", "setup field key")
	status := fs.String("status", "", "setup field status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *key == "" {
		return errors.New("setup field key is required")
	}
	if *status == "" {
		return errors.New("setup field status is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	setupStatus, err := supervisor.UpdateSetupField(ctx, *key, *status)
	if err != nil {
		return err
	}
	fmt.Println(setupStatus.Complete)
	return nil
}
