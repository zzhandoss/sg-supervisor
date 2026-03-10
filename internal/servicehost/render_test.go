package servicehost

import (
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/config"
)

func TestRenderWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	cfg := config.SupervisorConfig{
		ProductName:   "School Gate",
		ListenAddress: "127.0.0.1:8787",
	}

	plan := BuildPlan(layout, cfg, filepath.Join(root, "sg-supervisor.exe"))
	rendered, err := Render(plan)
	if err != nil {
		t.Fatalf("render service host artifacts: %v", err)
	}
	if len(rendered.WrittenFiles) != 6 {
		t.Fatalf("expected 6 written files, got %d", len(rendered.WrittenFiles))
	}
	for _, path := range rendered.WrittenFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s to exist: %v", path, err)
		}
	}
}
