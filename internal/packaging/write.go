package packaging

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sg-supervisor/internal/config"
)

func WriteManifest(layout config.Layout, manifest Manifest) (string, error) {
	outputPath := filepath.Join(layout.RuntimeDir, "packaging", manifest.Platform, "install-manifest.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return "", err
	}
	return outputPath, nil
}
