package app

import (
	"context"

	"sg-supervisor/internal/release"
)

func (a *App) BuildRelease(ctx context.Context, platform, version, binaryPath string) (release.Report, error) {
	distributionReport, err := a.BuildDistribution(ctx, platform, binaryPath)
	if err != nil {
		return release.Report{}, err
	}
	return release.Build(a.layout.Root, version, distributionReport)
}

func (a *App) BuildReleaseSet(ctx context.Context, version, binaryPath string) (release.SetReport, error) {
	platforms := []string{"windows", "linux"}
	reports := make([]release.Report, 0, len(platforms))
	for _, platform := range platforms {
		report, err := a.BuildRelease(ctx, platform, version, binaryPath)
		if err != nil {
			return release.SetReport{}, err
		}
		reports = append(reports, report)
	}
	return release.BuildSet(a.layout.Root, version, reports)
}
