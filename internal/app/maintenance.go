package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/maintenance"
	"sg-supervisor/internal/servicehost"
)

func (a *App) InstallPackage(ctx context.Context, packageID, binaryPath string) (control.InstallReport, error) {
	report := control.InstallReport{PackageID: packageID}
	active, err := a.ApplyPackage(ctx, packageID)
	if err != nil {
		return report, err
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
	if err := servicehost.ExecuteInstall(ctx, rendered.Plan, a.runner); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "service-registration",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("install completed with issues")
	}
	if err := a.refreshPackagingManifests(ctx, rendered, packageID, active.PackageID); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "refresh-packaging-manifests",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("install completed with issues")
	}
	report.Completed = true
	return report, nil
}

func (a *App) Repair(ctx context.Context, binaryPath string) (control.RepairReport, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.RepairReport{}, err
	}
	report := control.RepairReport{
		EnsuredPaths: []string{
			a.layout.InstallDir,
			a.layout.ConfigDir,
			a.layout.DataDir,
			a.layout.LogsDir,
			a.layout.LicensesDir,
			a.layout.BackupsDir,
			a.layout.RuntimeDir,
			a.layout.UpdatesDir,
			filepath.Join(a.layout.ConfigDir, "services.json"),
			filepath.Join(a.layout.ConfigDir, "supervisor.json"),
		},
	}
	rendered, err := a.renderServiceHost(ctx, binaryPath)
	if err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "render-service-host",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("repair completed with issues")
	}
	if err := servicehost.ExecuteRepair(ctx, rendered.Plan, a.runner); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "service-repair",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("repair completed with issues")
	}
	serviceHost := mapServiceHostArtifacts(rendered)
	report.ServiceArtifacts = serviceHost.WrittenFiles
	active, err := a.updates.Active(ctx)
	if err != nil {
		return report, err
	}
	report.ActivePackageID = active.PackageID
	report.NeedsPackageInstall = active.PackageID == ""
	if err := a.refreshPackagingManifests(ctx, rendered, active.PackageID, active.PackageID); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "refresh-packaging-manifests",
			Severity: "error",
			Message:  err.Error(),
		})
		return report, errors.New("repair completed with issues")
	}
	report.Completed = true
	return report, nil
}

func (a *App) Uninstall(ctx context.Context, mode string) (control.UninstallReport, error) {
	report := maintenance.UninstallReport{Mode: mode}
	rendered, err := a.renderServiceHost(ctx, "")
	if err != nil {
		report.Issues = append(report.Issues, maintenance.Issue{
			Step:     "render-service-host",
			Severity: "error",
			Message:  err.Error(),
		})
	}
	runningServices := a.runtime.RunningServiceNames()
	if len(runningServices) > 0 {
		if err := a.runtime.StopMany(runningServices); err != nil {
			report.Issues = append(report.Issues, maintenance.Issue{
				Step:     "stop-services",
				Severity: "error",
				Message:  err.Error(),
			})
		}
		if err := a.runtime.WaitForStopped(ctx, runningServices); err != nil {
			report.Issues = append(report.Issues, maintenance.Issue{
				Step:     "wait-for-stop",
				Severity: "error",
				Message:  err.Error(),
			})
		}
	}
	if err == nil {
		if uninstallErr := servicehost.ExecuteUninstall(ctx, rendered.Plan, a.runner); uninstallErr != nil {
			report.Issues = append(report.Issues, maintenance.Issue{
				Step:     "service-deregistration",
				Severity: "error",
				Message:  uninstallErr.Error(),
			})
		}
	}

	fsReport, uninstallErr := maintenance.ExecuteUninstall(a.layout, mode)
	report.Mode = fsReport.Mode
	report.Completed = fsReport.Completed
	report.RemovedPaths = fsReport.RemovedPaths
	report.KeptPaths = fsReport.KeptPaths
	report.Issues = append(report.Issues, fsReport.Issues...)
	if uninstallErr != nil {
		report.Issues = append(report.Issues, maintenance.Issue{
			Step:     "remove-managed-paths",
			Severity: "error",
			Message:  uninstallErr.Error(),
		})
	}
	report.StoppedServices = append(report.StoppedServices, runningServices...)
	if err == nil {
		report.UninstallHints = mapServiceHostArtifacts(rendered).UninstallHints
	}

	controlReport := control.UninstallReport{
		Mode:            report.Mode,
		Completed:       report.Completed,
		RemovedPaths:    report.RemovedPaths,
		KeptPaths:       report.KeptPaths,
		StoppedServices: report.StoppedServices,
		UninstallHints:  report.UninstallHints,
		Issues:          controlIssues(report.Issues),
	}
	if uninstallErr != nil {
		return controlReport, fmt.Errorf("uninstall failed with issues: %w", uninstallErr)
	}
	if len(report.Issues) > 0 {
		return controlReport, errors.New("uninstall completed with issues")
	}
	return controlReport, nil
}

func controlIssues(issues []maintenance.Issue) []control.Issue {
	result := make([]control.Issue, 0, len(issues))
	for _, issue := range issues {
		result = append(result, control.Issue{
			Step:     issue.Step,
			Severity: issue.Severity,
			Message:  issue.Message,
		})
	}
	return result
}
