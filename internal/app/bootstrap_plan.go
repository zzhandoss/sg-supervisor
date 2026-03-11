package app

import "path/filepath"

type bootstrapStep struct {
	Name    string
	Message string
	Args    []string
}

type bootstrapDeployTarget struct {
	Filter     string
	TargetPath string
}

func bootstrapBuildSteps() []bootstrapStep {
	return []bootstrapStep{
		{Name: "build-contracts", Message: "Building package @school-gate/contracts", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/contracts/tsconfig.json", "--noCheck"}},
		{Name: "build-core", Message: "Building package @school-gate/core", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/core/tsconfig.json", "--noCheck"}},
		{Name: "build-db", Message: "Building package @school-gate/db", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/db/tsconfig.json", "--noCheck"}},
		{Name: "build-config", Message: "Building package @school-gate/config", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/config/tsconfig.json", "--noCheck"}},
		{Name: "build-device", Message: "Building package @school-gate/device", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/device/tsconfig.json", "--noCheck"}},
		{Name: "build-infra", Message: "Building package @school-gate/infra", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/infra/tsconfig.json", "--noCheck"}},
		{Name: "build-ops", Message: "Building package @school-gate/ops", Args: []string{"pnpm", "exec", "tsc", "-p", "packages/ops/tsconfig.json", "--noCheck"}},
		{Name: "build-api", Message: "Building app @school-gate/api", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/api/tsconfig.json", "--noCheck"}},
		{Name: "build-device-service", Message: "Building app @school-gate/device-service", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/device-service/tsconfig.json", "--noCheck"}},
		{Name: "build-bot", Message: "Building app @school-gate/bot", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/bot/tsconfig.json", "--noCheck"}},
		{Name: "build-worker", Message: "Building app @school-gate/worker", Args: []string{"pnpm", "exec", "tsc", "-p", "apps/worker/tsconfig.json", "--noCheck"}},
		{Name: "build-admin-ui", Message: "Building app admin-ui", Args: []string{"pnpm", "--filter", "admin-ui", "build"}},
	}
}

func bootstrapDeployTargets() []bootstrapDeployTarget {
	return []bootstrapDeployTarget{
		{Filter: "@school-gate/ops", TargetPath: filepath.Join("packages", "ops")},
		{Filter: "@school-gate/api", TargetPath: filepath.Join("apps", "api")},
		{Filter: "@school-gate/device-service", TargetPath: filepath.Join("apps", "device-service")},
		{Filter: "@school-gate/bot", TargetPath: filepath.Join("apps", "bot")},
		{Filter: "@school-gate/worker", TargetPath: filepath.Join("apps", "worker")},
	}
}
