package app

import (
	"context"
	"path/filepath"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/packaging"
	"sg-supervisor/internal/servicehost"
)

func (a *App) refreshPackagingManifests(ctx context.Context, rendered servicehost.RenderedArtifacts, packageID, activePackageID string) error {
	if err := a.writePackagingManifest(rendered, packageID, activePackageID, "windows"); err != nil {
		return err
	}
	return a.writePackagingManifest(rendered, packageID, activePackageID, "linux")
}

func (a *App) writePackagingManifest(rendered servicehost.RenderedArtifacts, packageID, activePackageID, platform string) error {
	manifest, err := packaging.BuildManifest(a.layout, a.cfg, rendered, packageID, activePackageID, platform)
	if err != nil {
		return err
	}
	_, err = packaging.WriteManifest(a.layout, manifest)
	return err
}

func (a *App) PackagingStatus(ctx context.Context) (control.DirectoryStatus, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.DirectoryStatus{}, err
	}
	return control.DirectoryStatus{Runtime: a.layout.RuntimeDir}, nil
}

func (a *App) AssemblePackage(ctx context.Context, platform, binaryPath string) (packaging.AssembleReport, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return packaging.AssembleReport{}, err
	}
	rendered, err := a.renderServiceHost(ctx, binaryPath)
	if err != nil {
		return packaging.AssembleReport{}, err
	}
	active, err := a.updates.Active(ctx)
	if err != nil {
		return packaging.AssembleReport{}, err
	}
	if err := a.refreshPackagingManifests(ctx, rendered, active.PackageID, active.PackageID); err != nil {
		return packaging.AssembleReport{}, err
	}

	manifestPath := filepath.Join(a.layout.RuntimeDir, "packaging", platform, "install-manifest.json")
	manifest, err := packaging.LoadManifest(manifestPath)
	if err != nil {
		return packaging.AssembleReport{}, err
	}
	return packaging.Assemble(a.layout.Root, manifest)
}
