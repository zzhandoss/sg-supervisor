package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRuntimeDirectoriesCreatesServicePaths(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	if err := EnsureRuntimeDirectories(layout); err != nil {
		t.Fatalf("ensure runtime directories: %v", err)
	}

	required := []string{
		filepath.Join(layout.DataDir, "school-gate"),
		filepath.Join(layout.DataDir, "dahua-terminal-adapter"),
		filepath.Join(layout.LogsDir, "dahua-terminal-adapter"),
		filepath.Join(layout.RuntimeDir, "config"),
	}
	for _, path := range required {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s to exist", path)
		}
	}
}
