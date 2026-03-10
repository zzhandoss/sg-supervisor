package releasepanel

import (
	"encoding/json"
	"os"

	"sg-supervisor/internal/config"
)

type WorkspaceAssets struct {
	SchoolGateSourcePath string
	AdapterPath          string
	NodePath             string
	WinSWPath            string
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
