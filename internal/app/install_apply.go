package app

import (
	"context"
	"errors"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/servicehost"
)

func (a *App) installActivePackage(ctx context.Context, packageID, binaryPath string, registerService bool) (control.InstallReport, error) {
	report := control.InstallReport{PackageID: packageID}
	active, err := a.updates.Active(ctx)
	if err != nil {
		return report, err
	}
	if active.PackageID == "" {
		return report, errors.New("no active package after apply")
	}
	if active.PackageID != packageID {
		return report, errors.New("active package does not match requested install package")
	}
	report.ActivePackageID = active.PackageID
	rendered, err := a.renderServiceHost(ctx, binaryPath)
	if err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "render-service-host",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("install completed with issues")
	}
	serviceHost := mapServiceHostArtifacts(rendered)
	report.ServiceName = serviceHost.ServiceName
	report.WrittenFiles = serviceHost.WrittenFiles
	report.InstallHints = serviceHost.InstallHints
	if registerService {
		if err := a.executeInstallSideEffects(ctx, rendered, &report); err != nil {
			return report, err
		}
	}
	report.Completed = true
	return report, nil
}

func (a *App) executeInstallSideEffects(ctx context.Context, rendered servicehost.RenderedArtifacts, report *control.InstallReport) error {
	if err := servicehost.ExecuteInstall(ctx, rendered.Plan, a.runner); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "service-registration",
			Severity: "error",
			Message:  err.Error(),
		})
		return errors.New("install completed with issues")
	}
	return nil
}
