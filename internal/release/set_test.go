package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSetWritesAggregateMetadata(t *testing.T) {
	root := t.TempDir()
	releaseDir := filepath.Join(root, "releases", "v1.2.3", "linux")
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		t.Fatalf("mkdir release: %v", err)
	}

	report, err := BuildSet(root, "1.2.3", []Report{
		{Version: "1.2.3", Platform: "windows", ArtifactPath: filepath.Join(root, "releases", "v1.2.3", "windows", "a.msi")},
		{Version: "1.2.3", Platform: "linux", ArtifactPath: filepath.Join(root, "releases", "v1.2.3", "linux", "a.tar.gz")},
	})
	if err != nil {
		t.Fatalf("build set: %v", err)
	}
	if _, err := os.Stat(report.MetadataPath); err != nil {
		t.Fatalf("expected metadata path: %v", err)
	}
	if len(report.Platforms) != 2 || report.Platforms[0] != "linux" {
		t.Fatalf("expected sorted platforms, got %+v", report.Platforms)
	}
}
