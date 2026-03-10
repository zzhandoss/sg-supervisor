package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureServiceCatalogMigratesLegacyAdminUI(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	legacy := `{
  "services": [
    {
      "name": "admin-ui",
      "kind": "static-assets",
      "staticDir": "C:\\legacy\\admin-ui\\dist",
      "requiresLicense": true
    }
  ]
}
`
	if err := os.WriteFile(ServicesFile(layout), []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy catalog: %v", err)
	}

	if err := EnsureServiceCatalog(layout); err != nil {
		t.Fatalf("ensure service catalog: %v", err)
	}
	data, err := os.ReadFile(ServicesFile(layout))
	if err != nil {
		t.Fatalf("read catalog: %v", err)
	}
	if !strings.Contains(string(data), `"kind": "process-group"`) {
		t.Fatalf("expected migrated admin-ui process-group, got %s", string(data))
	}
	if !strings.Contains(string(data), `.output\\server\\index.mjs`) && !strings.Contains(string(data), `.output/server/index.mjs`) {
		t.Fatalf("expected admin-ui command in migrated catalog, got %s", string(data))
	}
	if runtime.GOOS == "windows" && !strings.Contains(string(data), filepath.Base(filepath.Join(root, "install", "runtime", "node", "node.exe"))) {
		t.Fatalf("expected node executable in migrated catalog, got %s", string(data))
	}
}
