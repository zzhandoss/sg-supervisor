package app

import (
	"context"
	"errors"
	"os"
	"time"

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

func (a *App) buildServiceHostPlan(ctx context.Context, binaryPath string) (servicehost.Plan, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return servicehost.Plan{}, err
	}
	if binaryPath == "" {
		current, err := os.Executable()
		if err != nil {
			return servicehost.Plan{}, err
		}
		binaryPath = current
	}
	return servicehost.BuildPlan(a.layout, a.cfg, binaryPath), nil
}

func (a *App) renderServiceHost(ctx context.Context, binaryPath string) (servicehost.RenderedArtifacts, error) {
	plan, err := a.buildServiceHostPlan(ctx, binaryPath)
	if err != nil {
		return servicehost.RenderedArtifacts{}, err
	}
	return servicehost.Render(plan)
}

func mapServiceHostArtifacts(rendered servicehost.RenderedArtifacts) control.ServiceHostArtifacts {
	return control.ServiceHostArtifacts{
		ServiceName:          rendered.Plan.ServiceName,
		DisplayName:          rendered.Plan.DisplayName,
		Description:          rendered.Plan.Description,
		BinaryPath:           rendered.Plan.BinaryPath,
		Arguments:            rendered.Plan.Arguments,
		ListenAddress:        rendered.Plan.ListenAddress,
		WindowsWrapperPath:   rendered.Plan.WindowsWrapperPath,
		WindowsConfigPath:    rendered.Plan.WindowsConfigPath,
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

func (a *App) ServiceHostStatus(ctx context.Context) (control.ServiceHostStatus, error) {
	plan, err := a.buildServiceHostPlan(ctx, "")
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	status, err := servicehost.QueryStatus(ctx, plan)
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	return mapServiceHostStatus(status), nil
}

func (a *App) InstallServiceHost(ctx context.Context) (control.ServiceHostStatus, error) {
	rendered, err := a.renderServiceHost(ctx, "")
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	if err := servicehost.InstallService(ctx, rendered.Plan); err != nil {
		status, _ := servicehost.QueryStatus(ctx, rendered.Plan)
		return mapServiceHostStatus(status), err
	}
	return a.ServiceHostStatus(ctx)
}

func (a *App) StartServiceHost(ctx context.Context) (control.ServiceHostStatus, error) {
	if a.isServing() {
		status, err := a.ServiceHostStatus(ctx)
		if err != nil {
			return control.ServiceHostStatus{}, err
		}
		return status, errors.New("Control Panel is already running interactively on this machine; use \"Switch to service mode\" instead of starting the Windows service directly")
	}
	return a.mutateServiceHost(ctx, servicehost.StartService)
}

func (a *App) SwitchToServiceHost(ctx context.Context) (control.ServiceHostStatus, error) {
	rendered, err := a.renderServiceHost(ctx, "")
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	status, err := servicehost.QueryStatus(ctx, rendered.Plan)
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	if !status.Installed {
		if err := servicehost.InstallService(ctx, rendered.Plan); err != nil {
			status, _ = servicehost.QueryStatus(ctx, rendered.Plan)
			return mapServiceHostStatus(status), err
		}
	}
	if err := servicehost.ScheduleStart(ctx, rendered.Plan, 2*time.Second); err != nil {
		status, _ = servicehost.QueryStatus(ctx, rendered.Plan)
		return mapServiceHostStatus(status), err
	}
	if !a.stopCurrentServe() {
		return mapServiceHostStatus(servicehost.Status{
			Supported:   status.Supported,
			ServiceName: rendered.Plan.ServiceName,
			Installed:   true,
			State:       "scheduled",
			StartMode:   status.StartMode,
			WrapperPath: rendered.Plan.WindowsWrapperPath,
			ConfigPath:  rendered.Plan.WindowsConfigPath,
			Description: rendered.Plan.Description,
		}), errors.New("service start was scheduled, but this instance is not running the active Control Panel host")
	}
	return control.ServiceHostStatus{
		Supported:   true,
		ServiceName: rendered.Plan.ServiceName,
		Installed:   true,
		State:       "switching",
		StartMode:   status.StartMode,
		WrapperPath: rendered.Plan.WindowsWrapperPath,
		ConfigPath:  rendered.Plan.WindowsConfigPath,
		Description: rendered.Plan.Description,
	}, nil
}

func (a *App) StopServiceHost(ctx context.Context) (control.ServiceHostStatus, error) {
	return a.mutateServiceHost(ctx, servicehost.StopService)
}

func (a *App) EnableServiceHostAutostart(ctx context.Context) (control.ServiceHostStatus, error) {
	return a.mutateServiceHost(ctx, servicehost.EnableAutostart)
}

func (a *App) DisableServiceHostAutostart(ctx context.Context) (control.ServiceHostStatus, error) {
	return a.mutateServiceHost(ctx, servicehost.DisableAutostart)
}

func (a *App) RemoveServiceHost(ctx context.Context) (control.ServiceHostStatus, error) {
	return a.mutateServiceHost(ctx, servicehost.RemoveService)
}

func (a *App) mutateServiceHost(ctx context.Context, action func(context.Context, servicehost.Plan) error) (control.ServiceHostStatus, error) {
	rendered, err := a.renderServiceHost(ctx, "")
	if err != nil {
		return control.ServiceHostStatus{}, err
	}
	if err := action(ctx, rendered.Plan); err != nil {
		status, _ := servicehost.QueryStatus(ctx, rendered.Plan)
		return mapServiceHostStatus(status), err
	}
	return a.ServiceHostStatus(ctx)
}

func mapServiceHostStatus(status servicehost.Status) control.ServiceHostStatus {
	return control.ServiceHostStatus{
		Supported:   status.Supported,
		ServiceName: status.ServiceName,
		Installed:   status.Installed,
		State:       status.State,
		StartMode:   status.StartMode,
		WrapperPath: status.WrapperPath,
		ConfigPath:  status.ConfigPath,
		LastError:   status.LastError,
		Description: status.Description,
	}
}
