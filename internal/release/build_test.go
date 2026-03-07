package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sg-supervisor/internal/distribution"
)

func TestBuildLinuxRelease(t *testing.T) {
	root := t.TempDir()
	distDir := filepath.Join(root, "dist", "linux")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatalf("mkdir dist: %v", err)
	}
	artifactPath := filepath.Join(distDir, "school-gate-linux-x64.tar.gz")
	if err := os.WriteFile(artifactPath, []byte("artifact"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	report, err := Build(root, "1.2.3", distribution.Report{
		Platform:     "linux",
		StageDir:     filepath.Join(root, "build", "linux"),
		OutputDir:    distDir,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		t.Fatalf("build release: %v", err)
	}
	if !strings.HasSuffix(report.ArtifactPath, "school-gate-installer-v1.2.3-linux-x64.tar.gz") {
		t.Fatalf("unexpected artifact path: %s", report.ArtifactPath)
	}
	if !strings.HasSuffix(report.ChecksumsPath, "school-gate-installer-v1.2.3-linux-x64-SHA256SUMS.txt") {
		t.Fatalf("unexpected checksums path: %s", report.ChecksumsPath)
	}
	if !strings.HasSuffix(report.MetadataPath, "school-gate-installer-v1.2.3-linux-x64-release.json") {
		t.Fatalf("unexpected metadata path: %s", report.MetadataPath)
	}
	if _, err := os.Stat(report.ChecksumsPath); err != nil {
		t.Fatalf("expected checksums file: %v", err)
	}
}

func TestBuildWindowsReleaseFallsBackToInputsZip(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "dist", "windows")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir dist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "Product.wxs"), []byte("wix"), 0o644); err != nil {
		t.Fatalf("write wix source: %v", err)
	}

	report, err := Build(root, "1.2.3", distribution.Report{
		Platform:  "windows",
		StageDir:  filepath.Join(root, "build", "windows"),
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("build release: %v", err)
	}
	if !strings.HasSuffix(report.ArtifactPath, "school-gate-installer-v1.2.3-windows-x64.zip") {
		t.Fatalf("unexpected artifact path: %s", report.ArtifactPath)
	}
	if !strings.HasSuffix(report.ChecksumsPath, "school-gate-installer-v1.2.3-windows-x64-SHA256SUMS.txt") {
		t.Fatalf("unexpected checksums path: %s", report.ChecksumsPath)
	}
	if !strings.HasSuffix(report.MetadataPath, "school-gate-installer-v1.2.3-windows-x64-release.json") {
		t.Fatalf("unexpected metadata path: %s", report.MetadataPath)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected fallback warning")
	}
}
