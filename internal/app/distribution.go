package app

import (
	"context"

	"sg-supervisor/internal/distribution"
)

func (a *App) BuildDistribution(ctx context.Context, platform, binaryPath string) (distribution.Report, error) {
	stage, err := a.AssemblePackage(ctx, platform, binaryPath)
	if err != nil {
		return distribution.Report{}, err
	}
	return distribution.Build(a.layout.Root, stage)
}
