package app

import (
	"context"
	"errors"
	"strings"

	"sg-supervisor/internal/updates"
)

func (a *App) ApplyLocalBundle(ctx context.Context, path string) (updates.ActiveRecord, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return updates.ActiveRecord{}, err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return updates.ActiveRecord{}, errors.New("path is required")
	}
	record, err := a.updates.ImportBundle(ctx, path)
	if err != nil {
		return updates.ActiveRecord{}, err
	}
	return a.ApplyPackage(ctx, record.PackageID)
}
