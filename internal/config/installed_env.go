package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func SchoolGateEnvFile(layout Layout) string {
	return filepath.Join(layout.InstallDir, "core", ".env")
}

func AdapterEnvFile(layout Layout) string {
	return filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", ".env")
}

func WriteInstalledEnvFiles(layout Layout, catalog ServiceCatalog) error {
	coreEnv := map[string]string{}
	adapterEnv := map[string]string{}
	for _, service := range catalog.Services {
		switch service.Name {
		case "api", "device-service", "bot", "worker", "admin-ui":
			mergeInstalledEnv(coreEnv, service.Env)
		case "dahua-terminal-adapter":
			mergeInstalledEnv(adapterEnv, service.Env)
		}
	}
	if err := writeEnvFile(SchoolGateEnvFile(layout), coreEnv); err != nil {
		return err
	}
	if len(adapterEnv) == 0 {
		return nil
	}
	return writeEnvFile(AdapterEnvFile(layout), adapterEnv)
}

func mergeInstalledEnv(target, source map[string]string) {
	for key, value := range source {
		if strings.TrimSpace(value) == "" {
			continue
		}
		target[key] = value
	}
}

func writeEnvFile(path string, env map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+env[key])
	}
	data := strings.Join(lines, "\n")
	if data != "" {
		data += "\n"
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
