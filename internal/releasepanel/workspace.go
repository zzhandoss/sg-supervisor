package releasepanel

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"sg-supervisor/internal/config"
)

type WorkspaceAssets struct {
	CorePath    string
	AdapterPath string
	NodePath    string
}

func prepareWorkspace(root, platform string, state State, assets WorkspaceAssets) error {
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		return err
	}
	if err := extractArchive(assets.CorePath, filepath.Join(layout.InstallDir, "core")); err != nil {
		return err
	}
	if err := extractArchive(assets.AdapterPath, filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter")); err != nil {
		return err
	}
	if err := extractArchive(assets.NodePath, filepath.Join(layout.InstallDir, "runtime", "node")); err != nil {
		return err
	}
	if err := writeSupervisorConfig(layout.ConfigFile, state); err != nil {
		return err
	}
	return validateWorkspace(layout, platform)
}

func writeSupervisorConfig(path string, state State) error {
	cfg := config.SupervisorConfig{
		ProductName:                   "School Gate",
		ListenAddress:                 "0.0.0.0:8787",
		PublicKeyBase64:               state.Keys.LicensePublicKeyBase64,
		PackageSigningPublicKeyBase64: state.Keys.PackagePublicKeyBase64,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func validateWorkspace(layout config.Layout, platform string) error {
	required := []string{
		filepath.Join(layout.InstallDir, "core", "apps", "api", "dist", "index.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "device-service", "dist", "api", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "device-service", "dist", "outbox", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "bot", "dist", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "worker", "dist", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "worker", "dist", "accessEvents", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "worker", "dist", "outbox", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "worker", "dist", "retention", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "worker", "dist", "monitoring", "main.js"),
		filepath.Join(layout.InstallDir, "core", "apps", "admin-ui"),
		filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src", "index.js"),
	}
	if platform == "windows" {
		required = append(required, filepath.Join(layout.InstallDir, "runtime", "node", "node.exe"))
	} else {
		required = append(required, filepath.Join(layout.InstallDir, "runtime", "node", "bin", "node"))
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			return errors.New("workspace is missing required path: " + path)
		}
	}
	return nil
}
