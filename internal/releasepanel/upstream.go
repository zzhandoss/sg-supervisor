package releasepanel

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type AssetSpec struct {
	Repo    string
	Tag     string
	Pattern string
}

type AssetSource interface {
	ListVersions(ctx context.Context, repo string) ([]ReleaseVersion, error)
	DownloadReleaseAsset(ctx context.Context, spec AssetSpec, targetDir string) (string, error)
}

type GitHubAssetSource struct {
	executor Executor
}

func NewGitHubAssetSource(executor Executor) *GitHubAssetSource {
	return &GitHubAssetSource{executor: executor}
}

func (s *GitHubAssetSource) ListVersions(ctx context.Context, repo string) ([]ReleaseVersion, error) {
	output, err := s.executor.Run(ctx, ".", nil, "gh", "api", "repos/"+repo+"/releases?per_page=20")
	if err != nil {
		return nil, err
	}
	var payload []struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, err
	}
	versions := make([]ReleaseVersion, 0, len(payload))
	for _, item := range payload {
		versions = append(versions, ReleaseVersion{
			Tag:         item.TagName,
			Name:        item.Name,
			PublishedAt: item.PublishedAt,
		})
	}
	return versions, nil
}

func (s *GitHubAssetSource) DownloadReleaseAsset(ctx context.Context, spec AssetSpec, targetDir string) (string, error) {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	if _, err := s.executor.Run(ctx, targetDir, nil, "gh", "release", "download", spec.Tag, "--repo", spec.Repo, "--pattern", spec.Pattern, "--dir", targetDir, "--clobber"); err != nil {
		return "", err
	}
	matches, err := filepath.Glob(filepath.Join(targetDir, spec.Pattern))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", errors.New("expected exactly one downloaded asset for pattern " + spec.Pattern)
	}
	return matches[0], nil
}

func normalizeTag(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "v") {
		return value
	}
	return "v" + value
}
