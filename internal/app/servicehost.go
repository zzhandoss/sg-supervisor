package app

import (
	"context"
	"os"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/servicehost"
)

func (a *App) RenderServiceHostArtifacts(ctx context.Context, binaryPath string) (control.ServiceHostArtifacts, error) {
	rendered, err := a.renderServiceHost(ctx, binaryPath)
	if err != nil {
		return control.ServiceHostArtifacts{}, err
	}
	return mapServiceHostArtifacts(rendered), nil
}

func (a *App) renderServiceHost(ctx context.Context, binaryPath string) (servicehost.RenderedArtifacts, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return servicehost.RenderedArtifacts{}, err
	}
	if binaryPath == "" {
		current, err := os.Executable()
		if err != nil {
			return servicehost.RenderedArtifacts{}, err
		}
		binaryPath = current
	}
	return servicehost.Render(servicehost.BuildPlan(a.layout, a.cfg, binaryPath))
}

func mapServiceHostArtifacts(rendered servicehost.RenderedArtifacts) control.ServiceHostArtifacts {
	return control.ServiceHostArtifacts{
		ServiceName:          rendered.Plan.ServiceName,
		DisplayName:          rendered.Plan.DisplayName,
		Description:          rendered.Plan.Description,
		BinaryPath:           rendered.Plan.BinaryPath,
		Arguments:            rendered.Plan.Arguments,
		ListenAddress:        rendered.Plan.ListenAddress,
		LinuxUnitPath:        rendered.Plan.LinuxUnitPath,
		WindowsInstallPath:   rendered.Plan.WindowsInstallPath,
		WindowsUninstallPath: rendered.Plan.WindowsUninstallPath,
		WindowsStartPath:     rendered.Plan.WindowsStartPath,
		WindowsStopPath:      rendered.Plan.WindowsStopPath,
		WrittenFiles:         rendered.WrittenFiles,
		InstallHints:         rendered.InstallHints,
		UninstallHints:       rendered.UninstallHints,
	}
}
