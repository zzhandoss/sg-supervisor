package distribution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareWindowsBuildStageCopiesTree(t *testing.T) {
	sourceDir := filepath.Join(t.TempDir(), "build", "windows")
	sourceFile := filepath.Join(sourceDir, "install", "app", "file.txt")
	if err := os.MkdirAll(filepath.Dir(sourceFile), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(sourceFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stageDir, cleanup, err := prepareWindowsBuildStage(sourceDir)
	if err != nil {
		t.Fatalf("prepare build stage: %v", err)
	}
	defer cleanup()

	targetFile := filepath.Join(stageDir, "install", "app", "file.txt")
	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read staged file: %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("unexpected staged content: %q", string(data))
	}
}
