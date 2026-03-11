package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func writeInstalledRootPackageJSON(sourceRoot, coreRoot string) error {
	sourcePath := filepath.Join(sourceRoot, "package.json")
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}

	scripts, _ := body["scripts"].(map[string]any)
	if scripts == nil {
		scripts = map[string]any{}
	}
	backupArgs := "--root-dir ../.. --env-path .env"
	scripts["backup:create"] = "node packages/ops/dist/cli.js create " + backupArgs
	scripts["backup:verify"] = "node packages/ops/dist/cli.js verify " + backupArgs
	scripts["backup:prune"] = "node packages/ops/dist/cli.js prune " + backupArgs
	scripts["backup:restore"] = "node packages/ops/dist/cli.js restore " + backupArgs
	body["scripts"] = scripts

	encoded, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(filepath.Join(coreRoot, "package.json"), encoded, 0o644)
}
