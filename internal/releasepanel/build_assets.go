package releasepanel

import (
	"context"
	"path/filepath"

	"sg-supervisor/internal/release"
)

func (s *Service) downloadAssets(ctx context.Context, state State, platform string) (WorkspaceAssets, error) {
	cacheDir := filepath.Join(s.layout.CacheDir, "downloads")
	adapterPath, err := s.assets.DownloadReleaseAsset(ctx, AssetSpec{
		Repo:    RepoAdapter,
		Tag:     normalizeTag(state.Recipe.AdapterVersion),
		Pattern: adapterAssetPattern(state.Recipe.AdapterVersion, platform),
	}, filepath.Join(cacheDir, "adapter", trimVersion(state.Recipe.AdapterVersion), platform))
	if err != nil {
		return WorkspaceAssets{}, err
	}
	nodePath, err := s.node.Download(state.Recipe.NodeVersion, platform, filepath.Join(cacheDir, "node", trimVersion(state.Recipe.NodeVersion), platform))
	if err != nil {
		return WorkspaceAssets{}, err
	}
	return WorkspaceAssets{AdapterPath: adapterPath, NodePath: nodePath}, nil
}

func copyReleaseReport(root, version string, report release.Report) (release.Report, error) {
	targetDir := filepath.Join(root, "releases", "v"+trimVersion(version), report.Platform)
	if err := copyDir(report.ReleaseDir, targetDir); err != nil {
		return release.Report{}, err
	}
	report.ReleaseDir = targetDir
	report.ArtifactPath = filepath.Join(targetDir, filepath.Base(report.ArtifactPath))
	report.MetadataPath = filepath.Join(targetDir, filepath.Base(report.MetadataPath))
	report.ChecksumsPath = filepath.Join(targetDir, filepath.Base(report.ChecksumsPath))
	return report, nil
}

func adapterAssetPattern(version, platform string) string {
	base := "dahua-adapter-v" + trimVersion(version)
	if platform == "windows" {
		return base + "-win-x64.zip"
	}
	return base + "-linux-x64.zip"
}

func trimVersion(version string) string {
	version = normalizeTag(version)
	return version[1:]
}
