package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAdminUIBuildDirPrefersOutputPublic(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "admin-ui")
	outputDir := filepath.Join(appDir, ".output")
	outputPublicDir := filepath.Join(outputDir, "public")
	outputServerDir := filepath.Join(outputDir, "server")
	distDir := filepath.Join(appDir, "dist")
	if err := os.MkdirAll(outputPublicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputServerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputServerDir, "index.mjs"), []byte("server"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte("dist"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveAdminUIBuildDir(appDir)
	if err != nil {
		t.Fatal(err)
	}
	if got != outputDir {
		t.Fatalf("expected %q, got %q", outputDir, got)
	}
}

func TestResolveAdminUIBuildDirFallsBackToDist(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "admin-ui")
	distDir := filepath.Join(appDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte("dist"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveAdminUIBuildDir(appDir)
	if err != nil {
		t.Fatal(err)
	}
	if got != distDir {
		t.Fatalf("expected %q, got %q", distDir, got)
	}
}
