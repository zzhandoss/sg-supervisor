package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteInstalledRootPackageJSONPatchesBackupScripts(t *testing.T) {
	t.Parallel()

	sourceRoot := t.TempDir()
	coreRoot := t.TempDir()
	sourceBody := map[string]any{
		"name": "school-gate",
		"scripts": map[string]any{
			"backup:create": "node packages/ops/dist/cli.js create",
		},
	}
	data, err := json.Marshal(sourceBody)
	if err != nil {
		t.Fatalf("marshal source package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "package.json"), data, 0o644); err != nil {
		t.Fatalf("write source package.json: %v", err)
	}

	if err := writeInstalledRootPackageJSON(sourceRoot, coreRoot); err != nil {
		t.Fatalf("write installed root package.json: %v", err)
	}

	written, err := os.ReadFile(filepath.Join(coreRoot, "package.json"))
	if err != nil {
		t.Fatalf("read installed package.json: %v", err)
	}
	var body struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(written, &body); err != nil {
		t.Fatalf("unmarshal installed package.json: %v", err)
	}

	want := "node packages/ops/dist/cli.js create --root-dir ../.. --env-path .env"
	if body.Scripts["backup:create"] != want {
		t.Fatalf("unexpected backup:create script: %q", body.Scripts["backup:create"])
	}
	if body.Scripts["backup:verify"] == "" || body.Scripts["backup:prune"] == "" || body.Scripts["backup:restore"] == "" {
		t.Fatalf("backup scripts were not fully patched: %#v", body.Scripts)
	}
}
