package releasepanel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadReleaseAssetUsesTargetDirDirectly(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "downloads")
	executor := &fakeDownloadExecutor{}
	source := NewGitHubAssetSource(executor)

	path, err := source.DownloadReleaseAsset(context.Background(), AssetSpec{
		Repo:    RepoSchoolGate,
		Tag:     "v1.2.0",
		Pattern: "school-gate-v1.2.0-prebuilt.zip",
	}, targetDir)
	if err != nil {
		t.Fatalf("download asset: %v", err)
	}
	if filepath.Dir(path) != targetDir {
		t.Fatalf("expected asset in %s, got %s", targetDir, path)
	}
	if executor.commandDir != targetDir {
		t.Fatalf("expected command dir %s, got %s", targetDir, executor.commandDir)
	}
	if executor.downloadDir != targetDir {
		t.Fatalf("expected --dir %s, got %s", targetDir, executor.downloadDir)
	}
}

type fakeDownloadExecutor struct {
	commandDir  string
	downloadDir string
}

func (f *fakeDownloadExecutor) Run(_ context.Context, dir string, _ map[string]string, _ string, args ...string) ([]byte, error) {
	f.commandDir = dir
	for index := 0; index < len(args)-1; index++ {
		if args[index] == "--dir" {
			f.downloadDir = args[index+1]
		}
	}
	path := filepath.Join(f.downloadDir, "school-gate-v1.2.0-prebuilt.zip")
	if err := os.WriteFile(path, []byte("asset"), 0o644); err != nil {
		return nil, err
	}
	return nil, nil
}
