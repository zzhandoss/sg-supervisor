package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Layout struct {
	Root        string
	InstallDir  string
	ConfigDir   string
	DataDir     string
	LogsDir     string
	LicensesDir string
	BackupsDir  string
	RuntimeDir  string
	UpdatesDir  string
	ConfigFile  string
}

type SupervisorConfig struct {
	ProductName                   string `json:"productName"`
	ListenAddress                 string `json:"listenAddress"`
	PublicKeyBase64               string `json:"publicKeyBase64"`
	PackageSigningPublicKeyBase64 string `json:"packageSigningPublicKeyBase64"`
}

func NewLayout(root string) Layout {
	return Layout{
		Root:        root,
		InstallDir:  filepath.Join(root, "install"),
		ConfigDir:   filepath.Join(root, "config"),
		DataDir:     filepath.Join(root, "data"),
		LogsDir:     filepath.Join(root, "logs"),
		LicensesDir: filepath.Join(root, "licenses"),
		BackupsDir:  filepath.Join(root, "backups"),
		RuntimeDir:  filepath.Join(root, "runtime"),
		UpdatesDir:  filepath.Join(root, "updates"),
		ConfigFile:  filepath.Join(root, "config", "supervisor.json"),
	}
}

func EnsureLayout(layout Layout) error {
	dirs := []string{
		layout.Root,
		layout.InstallDir,
		layout.ConfigDir,
		layout.DataDir,
		layout.LogsDir,
		layout.LicensesDir,
		layout.BackupsDir,
		layout.RuntimeDir,
		layout.UpdatesDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(layout.ConfigFile); os.IsNotExist(err) {
		return writeConfig(layout.ConfigFile, defaultConfig())
	} else if err != nil {
		return err
	}
	return nil
}

func LoadOrCreate(path string) (SupervisorConfig, error) {
	layout := NewLayout(filepath.Dir(filepath.Dir(path)))
	if err := EnsureLayout(layout); err != nil {
		return SupervisorConfig{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return SupervisorConfig{}, err
	}

	var cfg SupervisorConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return SupervisorConfig{}, err
	}
	if cfg.ProductName == "" {
		cfg.ProductName = defaultConfig().ProductName
	}
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = defaultConfig().ListenAddress
	}
	return cfg, nil
}

func writeConfig(path string, cfg SupervisorConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func defaultConfig() SupervisorConfig {
	return SupervisorConfig{
		ProductName:   "School Gate",
		ListenAddress: "0.0.0.0:8787",
	}
}
