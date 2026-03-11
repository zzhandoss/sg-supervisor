package app

import (
	"context"
	"os"
	"path/filepath"

	"sg-supervisor/internal/bootstrap"
)

func (a *App) deployBootstrapSource(ctx context.Context, status *bootstrap.Status, sourceRoot string) error {
	if err := a.markBootstrapStep(status, "deploy-school-gate", "running", "Preparing runnable school-gate layout"); err != nil {
		return err
	}
	coreRoot := filepath.Join(a.layout.InstallDir, "core")
	if err := resetBootstrapCoreLayout(coreRoot); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	if err := deployBootstrapTargets(ctx, a.root, sourceRoot, coreRoot, bootstrapDeployTargets()); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	if err := deployBootstrapRootFiles(sourceRoot, coreRoot); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	if err := deployBootstrapAdminUI(sourceRoot, coreRoot); err != nil {
		a.failBootstrap(*status, "deploy-school-gate", err)
		return err
	}
	return a.completeBootstrapStep(status, "deploy-school-gate", "Runnable school-gate layout prepared")
}

func resetBootstrapCoreLayout(coreRoot string) error {
	return os.RemoveAll(coreRoot)
}

func deployBootstrapTargets(ctx context.Context, supervisorRoot, sourceRoot, coreRoot string, targets []bootstrapDeployTarget) error {
	for _, target := range targets {
		targetDir := filepath.Join(coreRoot, target.TargetPath)
		if _, err := runBootstrapCommand(ctx, sourceRoot, bootstrapCommandEnv(supervisorRoot), corepackExecutablePath(supervisorRoot), "pnpm", "--filter", target.Filter, "deploy", "--prod", "--legacy", targetDir); err != nil {
			return err
		}
	}
	return nil
}

func deployBootstrapRootFiles(sourceRoot, coreRoot string) error {
	return writeInstalledRootPackageJSON(sourceRoot, coreRoot)
}

func deployBootstrapAdminUI(sourceRoot, coreRoot string) error {
	adminSourceDir := filepath.Join(sourceRoot, "apps", "admin-ui")
	adminTargetDir := filepath.Join(coreRoot, "apps", "admin-ui")
	adminBuildDir, err := resolveAdminUIBuildDir(adminSourceDir)
	if err != nil {
		return err
	}
	if err := copyDir(adminBuildDir, filepath.Join(adminTargetDir, ".output")); err != nil {
		return err
	}
	return copyFile(filepath.Join(adminSourceDir, "package.json"), filepath.Join(adminTargetDir, "package.json"))
}
