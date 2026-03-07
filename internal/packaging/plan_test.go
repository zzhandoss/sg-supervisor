package packaging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/servicehost"
)

func TestBuildAndWriteManifest(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	cfg := config.SupervisorConfig{
		ProductName:   "School Gate",
		ListenAddress: "127.0.0.1:8787",
	}
	plan := servicehost.BuildPlan(layout, cfg, filepath.Join(root, "sg-supervisor.exe"))
	rendered, err := servicehost.Render(plan)
	if err != nil {
		t.Fatalf("render service host: %v", err)
	}

	manifest, err := BuildManifest(layout, cfg, rendered, "pkg-1", "pkg-1", "windows")
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if manifest.Platform != "windows" {
		t.Fatalf("unexpected platform: %s", manifest.Platform)
	}
	if len(manifest.InstallActions) == 0 {
		t.Fatalf("expected install actions")
	}
	for _, entry := range manifest.Files {
		if entry.Kind == "runtime-root" {
			t.Fatalf("packaging manifest should not copy the entire runtime root anymore")
		}
	}
	for _, entry := range manifest.Files {
		if entry.Kind == "service-host-artifact" && filepath.Ext(entry.TargetPath) == ".service" {
			t.Fatalf("windows manifest should not include linux unit: %+v", entry)
		}
	}
	path, err := WriteManifest(layout, manifest)
	if err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var stored Manifest
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if stored.ActivePackageID != "pkg-1" {
		t.Fatalf("unexpected stored manifest: %+v", stored)
	}
}

func TestAssembleCopiesManifestFiles(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "source")
	if err := os.MkdirAll(filepath.Join(sourceDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "nested", "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	supervisorPath := filepath.Join(root, "sg-supervisor.exe")
	if err := os.WriteFile(supervisorPath, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	report, err := Assemble(root, Manifest{
		Platform: "windows",
		Files: []FileEntry{
			{SourcePath: supervisorPath, TargetPath: filepath.Join("supervisor", "sg-supervisor.exe"), Kind: "supervisor-binary"},
			{SourcePath: sourceDir, TargetPath: "install", Kind: "product-install-root"},
		},
	})
	if err != nil {
		t.Fatalf("assemble package: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.OutputDir, "install", "nested", "file.txt")); err != nil {
		t.Fatalf("expected copied install file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.OutputDir, "supervisor", "sg-supervisor.exe")); err != nil {
		t.Fatalf("expected copied supervisor binary: %v", err)
	}
}
