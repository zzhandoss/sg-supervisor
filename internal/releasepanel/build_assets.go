package releasepanel

import (
	"context"
	"path/filepath"
)

func (s *Service) downloadAssets(ctx context.Context, state State, platform string) (WorkspaceAssets, error) {
	cacheDir := filepath.Join(s.layout.CacheDir, "downloads")
	sourcePath, err := s.assets.DownloadReleaseAsset(ctx, schoolGateSourceSpec(state.Recipe), filepath.Join(cacheDir, "school-gate", trimVersion(state.Recipe.SchoolGateVersion), "source"))
	if err != nil {
		return WorkspaceAssets{}, err
	}
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
	return WorkspaceAssets{
		SchoolGateSourcePath: sourcePath,
		AdapterPath:          adapterPath,
		NodePath:             nodePath,
	}, nil
}

func adapterAssetPattern(version, platform string) string {
	base := "dahua-adapter-v" + trimVersion(version)
	if platform == "windows" {
		return base + "-win-x64.zip"
	}
	return base + "-linux-x64.zip"
}

func schoolGateSourceSpec(recipe Recipe) AssetSpec {
	version := trimVersion(recipe.SchoolGateVersion)
	return AssetSpec{
		Repo:    RepoSchoolGate,
		Tag:     normalizeTag(recipe.SchoolGateVersion),
		Pattern: "school-gate-v" + version + "-source.zip",
	}
}

func trimVersion(version string) string {
	version = normalizeTag(version)
	return version[1:]
}
