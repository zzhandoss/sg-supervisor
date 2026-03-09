package distribution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderWiXSourceSkipsDevArtifacts(t *testing.T) {
	stageDir := t.TempDir()
	files := map[string]string{
		"install/app/index.js":           "js",
		"install/app/index.d.ts":         "types",
		"install/app/index.js.map":       "map",
		"install/app/README.md":          "docs",
		"install/app/tests/spec.test.js": "test",
	}
	for relativePath, content := range files {
		fullPath := filepath.Join(stageDir, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir stage: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write stage file: %v", err)
		}
	}

	source := renderWiXSource(stageDir)
	if !strings.Contains(source, "index.js") {
		t.Fatalf("expected runtime file in wix source")
	}
	if !strings.Contains(source, "bootstrap-install") {
		t.Fatalf("expected wix source to include bootstrap install custom action")
	}
	if !strings.Contains(source, `[SourceDir]payload`) {
		t.Fatalf("expected wix source to reference payload source directory")
	}
	for _, fragment := range []string{"index.d.ts", "index.js.map", "README.md", "spec.test.js"} {
		if strings.Contains(source, fragment) {
			t.Fatalf("expected wix source to skip %s", fragment)
		}
	}
}
