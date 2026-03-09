package releasepanel

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/config"
)

func TestBuildLocalPayloadBundleCreatesManifestAndPayload(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "core", "apps", "api", "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "core", "apps", "api", "dist", "index.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src", "index.js"), []byte("console.log('adapter')"), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewStore(NewLayout(t.TempDir())).Ensure(".")
	if err != nil {
		t.Fatal(err)
	}
	state.Recipe = Recipe{
		InstallerVersion:  "1.0.0",
		SchoolGateVersion: "1.2.0",
		AdapterVersion:    "0.2.0",
		NodeVersion:       "20.19.0",
	}

	targetPath := filepath.Join(t.TempDir(), "school-gate-package.zip")
	if _, err := buildLocalPayloadBundle(context.Background(), root, state, targetPath); err != nil {
		t.Fatalf("build local payload bundle: %v", err)
	}

	reader, err := zip.OpenReader(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	seen := map[string]bool{}
	for _, file := range reader.File {
		seen[file.Name] = true
	}
	for _, name := range []string{
		"manifest.json",
		"manifest.sig",
		"payload/core/apps/api/dist/index.js",
		"payload/adapters/dahua-terminal-adapter/dist/src/index.js",
	} {
		if !seen[name] {
			t.Fatalf("expected bundle entry %s", name)
		}
	}
}
