package app

import (
	"context"
	"path/filepath"

	"sg-supervisor/internal/bootstrap"
)

func (a *App) migrateBootstrapDatabases(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "migrate-databases", "running", "Applying school-gate database migrations"); err != nil {
		return err
	}
	env := bootstrapCommandEnv(a.root)
	env["DB_FILE"] = filepath.Join(a.layout.DataDir, "school-gate", "app.db")
	env["DEVICE_DB_FILE"] = filepath.Join(a.layout.DataDir, "school-gate", "device.db")

	if _, err := runBootstrapCommand(ctx, sourceRoot, env, corepackExecutablePath(a.root), "pnpm", "db:migrate"); err != nil {
		a.failBootstrap(*status, "migrate-databases", err)
		return err
	}
	if _, err := runBootstrapCommand(ctx, sourceRoot, env, corepackExecutablePath(a.root), "pnpm", "device:db:migrate"); err != nil {
		a.failBootstrap(*status, "migrate-databases", err)
		return err
	}
	return a.completeBootstrapStep(status, "migrate-databases", "Database migrations applied")
}
