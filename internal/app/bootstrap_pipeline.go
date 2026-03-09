package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"sg-supervisor/internal/bootstrap"
)

func (a *App) BootstrapStatus(ctx context.Context) (bootstrap.Status, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return bootstrap.Status{}, err
	}
	return a.bootstrap.Load()
}

func (a *App) StartBootstrap(ctx context.Context) (bootstrap.Status, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return bootstrap.Status{}, err
	}
	a.bootstrapMu.Lock()
	defer a.bootstrapMu.Unlock()

	status, err := a.bootstrap.Load()
	if err != nil {
		return bootstrap.Status{}, err
	}
	if status.State == "running" {
		return bootstrap.Status{}, errors.New("bootstrap pipeline is already running")
	}
	status = bootstrap.Status{
		State: "running",
		Steps: []bootstrap.StepStatus{
			{Name: "locate-assets", State: "pending"},
			{Name: "extract-source", State: "pending"},
			{Name: "prepare-pnpm", State: "pending"},
			{Name: "install-dependencies", State: "pending"},
			{Name: "build-school-gate", State: "pending"},
			{Name: "migrate-databases", State: "pending"},
			{Name: "deploy-school-gate", State: "pending"},
			{Name: "prepare-adapter", State: "pending"},
			{Name: "cleanup-workspace", State: "pending"},
		},
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := a.bootstrap.Save(status); err != nil {
		return bootstrap.Status{}, err
	}
	go a.runBootstrapPipeline(context.Background())
	return status, nil
}

func (a *App) runBootstrapPipeline(ctx context.Context) {
	status, err := a.bootstrap.Load()
	if err != nil {
		return
	}
	assets, err := locateBootstrapAssets(a.root, a.bootstrap.Dir())
	if err != nil {
		a.failBootstrap(status, "locate-assets", err)
		return
	}
	status.SourceArchivePath = assets.SourceArchivePath
	status.AdapterArchivePath = assets.AdapterArchivePath
	status.SourceRoot = assets.SourceRoot
	if err := a.markBootstrapStep(&status, "locate-assets", "running", "Bootstrap assets found"); err != nil {
		return
	}
	if err := a.completeBootstrapStep(&status, "locate-assets", "Bootstrap assets found"); err != nil {
		return
	}
	if err := a.extractBootstrapSource(ctx, &status, assets); err != nil {
		return
	}
	if err := a.prepareBootstrapPnpm(ctx, &status, assets.SourceRoot); err != nil {
		return
	}
	if err := a.installBootstrapDependencies(ctx, &status, assets.SourceRoot); err != nil {
		return
	}
	if err := a.buildBootstrapSource(ctx, &status, assets.SourceRoot); err != nil {
		return
	}
	if err := a.migrateBootstrapDatabases(ctx, &status, assets.SourceRoot); err != nil {
		return
	}
	if err := a.deployBootstrapSource(ctx, &status, assets.SourceRoot); err != nil {
		return
	}
	if err := a.prepareBootstrapAdapter(ctx, &status, assets); err != nil {
		return
	}
	if err := a.cleanupBootstrapWorkspace(&status); err != nil {
		return
	}
	status.State = "succeeded"
	status.CurrentStep = ""
	status.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	status.Logs = append(status.Logs, "Bootstrap pipeline finished")
	_ = a.bootstrap.Save(status)
}

func (a *App) extractBootstrapSource(ctx context.Context, status *bootstrap.Status, assets bootstrapAssets) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := a.markBootstrapStep(status, "extract-source", "running", "Extracting school-gate source archive"); err != nil {
		return err
	}
	sourceRoot := assets.SourceRoot
	if err := os.RemoveAll(sourceRoot); err != nil {
		return err
	}
	if err := extractBootstrapArchive(assets.SourceArchivePath, sourceRoot); err != nil {
		a.failBootstrap(*status, "extract-source", err)
		return err
	}
	return a.completeBootstrapStep(status, "extract-source", "Source archive extracted")
}

func (a *App) prepareBootstrapPnpm(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "prepare-pnpm", "running", "Preparing pnpm with corepack"); err != nil {
		return err
	}
	packageManager, err := detectPackageManager(sourceRoot)
	if err != nil {
		a.failBootstrap(*status, "prepare-pnpm", err)
		return err
	}
	if _, err := runBootstrapCommand(ctx, sourceRoot, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "prepare", packageManager, "--activate"); err != nil {
		a.failBootstrap(*status, "prepare-pnpm", err)
		return err
	}
	return a.completeBootstrapStep(status, "prepare-pnpm", "pnpm prepared")
}

func (a *App) installBootstrapDependencies(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "install-dependencies", "running", "Installing school-gate dependencies"); err != nil {
		return err
	}
	if _, err := runBootstrapCommand(ctx, sourceRoot, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "pnpm", "install", "--frozen-lockfile", "--force", "--config.node-linker=hoisted"); err != nil {
		a.failBootstrap(*status, "install-dependencies", err)
		return err
	}
	return a.completeBootstrapStep(status, "install-dependencies", "Dependencies installed")
}

func (a *App) buildBootstrapSource(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "build-school-gate", "running", "Building school-gate applications"); err != nil {
		return err
	}
	for _, step := range bootstrapBuildSteps() {
		status.Logs = append(status.Logs, step.Message)
		_ = a.bootstrap.Save(*status)
		args := append([]string{"pnpm"}, step.Args[1:]...)
		if _, err := runBootstrapCommand(ctx, sourceRoot, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), args...); err != nil {
			a.failBootstrap(*status, "build-school-gate", err)
			return err
		}
	}
	return a.completeBootstrapStep(status, "build-school-gate", "School-gate build finished")
}

func (a *App) deployBootstrapSource(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "deploy-school-gate", "running", "Preparing runnable school-gate layout"); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(a.layout.InstallDir, "core")); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	for _, target := range bootstrapDeployTargets() {
		targetDir := filepath.Join(a.layout.InstallDir, "core", target.TargetPath)
		if _, err := runBootstrapCommand(ctx, sourceRoot, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "pnpm", "--filter", target.Filter, "deploy", "--prod", "--legacy", targetDir); err != nil {
			a.failBootstrap(*status, "deploy-school-gate", err)
			return err
		}
	}
	adminSourceDir := filepath.Join(sourceRoot, "apps", "admin-ui")
	adminTargetDir := filepath.Join(a.layout.InstallDir, "core", "apps", "admin-ui")
	if err := copyDir(filepath.Join(adminSourceDir, ".output", "public"), filepath.Join(adminTargetDir, "dist")); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	if err := copyFile(filepath.Join(adminSourceDir, "package.json"), filepath.Join(adminTargetDir, "package.json")); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	return a.completeBootstrapStep(status, "deploy-school-gate", "Runnable school-gate layout prepared")
}

func (a *App) prepareBootstrapAdapter(ctx context.Context, status *bootstrap.Status, assets bootstrapAssets) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := a.markBootstrapStep(status, "prepare-adapter", "running", "Extracting adapter runtime"); err != nil {
		return err
	}
	targetDir := filepath.Join(a.layout.InstallDir, "adapters", "dahua-terminal-adapter")
	if err := os.RemoveAll(targetDir); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if err := extractBootstrapArchive(assets.AdapterArchivePath, targetDir); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	packageManager, err := detectPackageManager(targetDir)
	if err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if err := os.RemoveAll(filepath.Join(targetDir, "node_modules")); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, "pnpm-workspace.yaml"), []byte("onlyBuiltDependencies:\n  - better-sqlite3\n"), 0644); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if _, err := runBootstrapCommand(ctx, targetDir, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "prepare", packageManager, "--activate"); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if _, err := runBootstrapCommand(ctx, targetDir, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "pnpm", "install", "--frozen-lockfile", "--prod", "--force"); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	if _, err := runBootstrapCommand(ctx, targetDir, bootstrapCommandEnv(a.root), corepackExecutablePath(a.root), "pnpm", "rebuild"); err != nil {
		a.failBootstrap(*status, "prepare-adapter", err)
		return err
	}
	return a.completeBootstrapStep(status, "prepare-adapter", "Adapter prepared")
}
