package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/servicehost"
)

func (a *App) BootstrapInstall(ctx context.Context, bundleDir, binaryPath string) (control.InstallReport, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.InstallReport{}, err
	}
	bundlePath, err := locateLocalBundle(bundleDir)
	if err != nil {
		return control.InstallReport{}, err
	}
	active, err := a.ApplyLocalBundle(ctx, bundlePath)
	if err != nil {
		return control.InstallReport{}, err
	}
	return a.installActivePackage(ctx, active.PackageID, binaryPath)
}

func (a *App) installActivePackage(ctx context.Context, packageID, binaryPath string) (control.InstallReport, error) {
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
	if err := a.executeInstallSideEffects(ctx, rendered, packageID, active.PackageID, &report); err != nil {
		return report, err
	}
	report.Completed = true
	return report, nil
}

func (a *App) executeInstallSideEffects(ctx context.Context, rendered servicehost.RenderedArtifacts, packageID, activePackageID string, report *control.InstallReport) error {
	if err := servicehost.ExecuteInstall(ctx, rendered.Plan, a.runner); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "service-registration",
			Severity: "error",
			Message:  err.Error(),
		})
		return errors.New("install completed with issues")
	}
	if err := a.refreshPackagingManifests(ctx, rendered, packageID, activePackageID); err != nil {
		report.Issues = append(report.Issues, control.Issue{
			Step:     "refresh-packaging-manifests",
			Severity: "error",
			Message:  err.Error(),
		})
		return errors.New("install completed with issues")
	}
	return nil
}

func locateLocalBundle(bundleDir string) (string, error) {
	bundleDir = strings.TrimSpace(bundleDir)
	if bundleDir == "" {
		return "", errors.New("bundle-dir is required")
	}
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return "", err
	}
	matches := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "school-gate-package-") {
			matches = append(matches, filepath.Join(bundleDir, entry.Name()))
		}
	}
	sort.Strings(matches)
	switch len(matches) {
	case 0:
		return "", errors.New("no local payload bundle was found in " + bundleDir)
	case 1:
		return matches[0], nil
	default:
		return "", errors.New("multiple local payload bundles were found in " + bundleDir)
	}
}
