package releasepanel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyMaterializedDirSkipsDevArtifacts(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := filepath.Join(t.TempDir(), "target")
	files := map[string]string{
		"dist/index.js":          "js",
		"dist/index.d.ts":        "types",
		"dist/index.js.map":      "map",
		"README.md":              "docs",
		"tests/runtime.test.js":  "test",
		"package.json":           "pkg",
		"node_modules/a/app.js":  "dep",
		"node_modules/a/app.map": "dep-map",
	}
	for relativePath, content := range files {
		fullPath := filepath.Join(sourceDir, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir source: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write source: %v", err)
		}
	}

	if err := copyMaterializedDir(sourceDir, targetDir); err != nil {
		t.Fatalf("copy materialized dir: %v", err)
	}

	assertExists(t, filepath.Join(targetDir, "dist", "index.js"))
	assertExists(t, filepath.Join(targetDir, "package.json"))
	assertExists(t, filepath.Join(targetDir, "node_modules", "a", "app.js"))
	assertMissing(t, filepath.Join(targetDir, "dist", "index.d.ts"))
	assertMissing(t, filepath.Join(targetDir, "dist", "index.js.map"))
	assertMissing(t, filepath.Join(targetDir, "README.md"))
	assertMissing(t, filepath.Join(targetDir, "tests"))
	assertMissing(t, filepath.Join(targetDir, "node_modules", "a", "app.map"))
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s (%v)", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path to be missing: %s", path)
	}
}
