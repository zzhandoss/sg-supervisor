package releasepanel

import (
	"context"
	"errors"
	"path/filepath"

	"sg-supervisor/internal/config"
)

type CoreBuilder interface {
	BuildInstallTree(ctx context.Context, recipe Recipe, cacheDir, workspaceRoot string, logf func(string)) error
}

type SchoolGateCoreBuilder struct {
	executor Executor
	assets   AssetSource
}

func NewSchoolGateCoreBuilder(executor Executor, assets AssetSource) *SchoolGateCoreBuilder {
	return &SchoolGateCoreBuilder{executor: executor, assets: assets}
}

func (b *SchoolGateCoreBuilder) BuildInstallTree(ctx context.Context, recipe Recipe, cacheDir, workspaceRoot string, logf func(string)) error {
	sourcePath, err := b.assets.DownloadReleaseAsset(ctx, schoolGateSourceSpec(recipe), filepath.Join(cacheDir, "school-gate", trimVersion(recipe.SchoolGateVersion), "source"))
	if err != nil {
		return err
	}
	sourceRoot := filepath.Join(workspaceRoot, "build", "school-gate-source")
	if err := extractArchive(sourcePath, sourceRoot); err != nil {
		return err
	}
	if logf != nil {
		logf("installing school-gate dependencies")
	}
	if _, err := b.executor.Run(ctx, sourceRoot, nil, "pnpm", "install", "--frozen-lockfile", "--force", "--config.node-linker=hoisted"); err != nil {
		return errors.New(commandErrorText(err))
	}
	if err := b.buildWorkspace(ctx, sourceRoot, logf); err != nil {
		return err
	}
	return b.assembleInstallTree(ctx, sourceRoot, filepath.Join(config.NewLayout(workspaceRoot).InstallDir, "core"), workspaceRoot, logf)
}

func schoolGateSourceSpec(recipe Recipe) AssetSpec {
	version := trimVersion(recipe.SchoolGateVersion)
	return AssetSpec{
		Repo:    RepoSchoolGate,
		Tag:     normalizeTag(recipe.SchoolGateVersion),
		Pattern: "school-gate-v" + version + "-source.zip",
	}
}

func commandErrorText(err error) string {
	if output := commandOutput(err); output != "" {
		return output
	}
	return err.Error()
}
