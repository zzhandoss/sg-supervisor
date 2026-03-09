package app

import (
	"context"
	"errors"
)

func (a *App) StartService(ctx context.Context, name string) error {
	status, err := a.license.Status(ctx)
	if err != nil {
		return err
	}
	if name == "dahua-terminal-adapter" && !a.isServiceRunning("device-service") {
		return errors.New("device-service must be running before the dahua adapter can start")
	}
	return a.runtime.Start(ctx, name, status.Valid)
}

func (a *App) isServiceRunning(name string) bool {
	for _, status := range a.runtime.Statuses() {
		if status.Name == name {
			return status.State == "running"
		}
	}
	return false
}
