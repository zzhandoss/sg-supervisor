package distribution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sg-supervisor/internal/packaging"
)

func TestBuildLinuxDistribution(t *testing.T) {
	stageDir := filepath.Join(t.TempDir(), "build", "linux")
	if err := os.MkdirAll(filepath.Join(stageDir, "install"), 0o755); err != nil {
		t.Fatalf("mkdir stage: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stageDir, "install", "app.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write stage file: %v", err)
	}

	report, err := Build(filepath.Dir(filepath.Dir(stageDir)), packaging.AssembleReport{
		Platform:  "linux",
		OutputDir: stageDir,
	})
	if err != nil {
		t.Fatalf("build linux distribution: %v", err)
	}
	if !strings.HasSuffix(report.ArtifactPath, ".tar.gz") {
		t.Fatalf("expected tar.gz artifact, got %+v", report)
	}
	if _, err := os.Stat(report.ArtifactPath); err != nil {
		t.Fatalf("expected linux artifact: %v", err)
	}
}

func TestBuildWindowsDistributionGeneratesInputs(t *testing.T) {
	stageDir := filepath.Join(t.TempDir(), "build", "windows")
	if err := os.MkdirAll(filepath.Join(stageDir, "supervisor"), 0o755); err != nil {
		t.Fatalf("mkdir stage: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stageDir, "supervisor", "sg-supervisor.exe"), []byte("bin"), 0o755); err != nil {
		t.Fatalf("write stage file: %v", err)
	}

	report, err := Build(filepath.Dir(filepath.Dir(stageDir)), packaging.AssembleReport{
		Platform:  "windows",
		OutputDir: stageDir,
	})
	if err != nil {
		t.Fatalf("build windows distribution: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.OutputDir, "Product.wxs")); err != nil {
		t.Fatalf("expected WiX source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.OutputDir, "build-msi.ps1")); err != nil {
		t.Fatalf("expected build script: %v", err)
	}
}
