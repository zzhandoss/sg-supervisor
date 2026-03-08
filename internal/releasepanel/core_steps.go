package releasepanel

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type schoolGateStep struct {
	Name string
	Args []string
}

type schoolGateDeployTarget struct {
	Filter     string
	TargetPath string
}

func (b *SchoolGateCoreBuilder) buildWorkspace(ctx context.Context, sourceRoot string, logf func(string)) error {
	for _, step := range schoolGateBuildSteps() {
		if logf != nil {
			logf(step.Name)
		}
		if _, err := b.executor.Run(ctx, sourceRoot, nil, step.Args[0], step.Args[1:]...); err != nil {
			return errors.New(commandErrorText(err))
		}
	}
	return nil
}

func (b *SchoolGateCoreBuilder) assembleInstallTree(ctx context.Context, sourceRoot, installCoreDir, workspaceRoot string, logf func(string)) error {
	deployRoot := filepath.Join(workspaceRoot, "build", "school-gate-deploy")
	for _, target := range schoolGateDeployTargets() {
		deployedDir, err := filepath.Abs(filepath.Join(deployRoot, target.TargetPath))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(deployedDir), 0o755); err != nil {
			return err
		}
		if logf != nil {
			logf("deploying " + target.Filter)
		}
		if _, err := b.executor.Run(ctx, sourceRoot, nil, "pnpm", "--filter", target.Filter, "deploy", "--prod", "--legacy", deployedDir); err != nil {
			return errors.New(commandErrorText(err))
		}
		if err := copyMaterializedDir(deployedDir, filepath.Join(installCoreDir, target.TargetPath)); err != nil {
			return err
		}
	}
	if logf != nil {
		logf("copying admin-ui build output")
	}
	adminSourceDir := filepath.Join(sourceRoot, "apps", "admin-ui")
	adminTargetDir := filepath.Join(installCoreDir, "apps", "admin-ui")
	if err := copyDir(filepath.Join(adminSourceDir, ".output", "public"), filepath.Join(adminTargetDir, "dist")); err != nil {
		return err
	}
	return copyFile(filepath.Join(adminSourceDir, "package.json"), filepath.Join(adminTargetDir, "package.json"))
}

func schoolGateBuildSteps() []schoolGateStep {
	return []schoolGateStep{
		{Name: "building package @school-gate/contracts", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/contracts/tsconfig.json", "--noCheck"}},
		{Name: "building package @school-gate/core", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/core/tsconfig.json", "--noCheck"}},
		{Name: "building package @school-gate/db", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/db/tsconfig.json", "--noCheck"}},
		{Name: "building package @school-gate/config", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/config/tsconfig.json", "--noCheck"}},
		{Name: "building package @school-gate/device", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/device/tsconfig.json", "--noCheck"}},
		{Name: "building package @school-gate/infra", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/infra/tsconfig.json", "--noCheck"}},
		{Name: "building app @school-gate/api", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/api/tsconfig.json", "--noCheck"}},
		{Name: "building app @school-gate/device-service", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/device-service/tsconfig.json", "--noCheck"}},
		{Name: "building app @school-gate/bot", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/bot/tsconfig.json", "--noCheck"}},
		{Name: "building app @school-gate/worker", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/worker/tsconfig.json", "--noCheck"}},
		{Name: "building app admin-ui", Args: []string{"pnpm", "--filter", "admin-ui", "build"}},
	}
}

func schoolGateDeployTargets() []schoolGateDeployTarget {
	return []schoolGateDeployTarget{
		{Filter: "@school-gate/api", TargetPath: filepath.Join("apps", "api")},
		{Filter: "@school-gate/device-service", TargetPath: filepath.Join("apps", "device-service")},
		{Filter: "@school-gate/bot", TargetPath: filepath.Join("apps", "bot")},
		{Filter: "@school-gate/worker", TargetPath: filepath.Join("apps", "worker")},
	}
}

func copyFile(sourcePath, targetPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, data, 0o644)
}
