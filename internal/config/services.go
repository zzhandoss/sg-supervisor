package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type CommandSpec struct {
	Name       string   `json:"name"`
	Executable string   `json:"executable"`
	Args       []string `json:"args,omitempty"`
	WorkingDir string   `json:"workingDir,omitempty"`
}

type ServiceSpec struct {
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Commands        []CommandSpec     `json:"commands,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	StaticDir       string            `json:"staticDir,omitempty"`
	HealthChecks    []HealthCheckSpec `json:"healthChecks,omitempty"`
	RequiresLicense bool              `json:"requiresLicense"`
}

type ServiceCatalog struct {
	Services []ServiceSpec `json:"services"`
}

type HealthCheckSpec struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	TimeoutMS int    `json:"timeoutMs,omitempty"`
}

func ServicesFile(layout Layout) string {
	return filepath.Join(layout.ConfigDir, "services.json")
}

func EnsureServiceCatalog(layout Layout) error {
	path := ServicesFile(layout)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return writeServiceCatalog(path, defaultServiceCatalog(layout))
	} else if err != nil {
		return err
	}
	return nil
}

func LoadServiceCatalog(layout Layout) (ServiceCatalog, error) {
	if err := EnsureServiceCatalog(layout); err != nil {
		return ServiceCatalog{}, err
	}
	data, err := os.ReadFile(ServicesFile(layout))
	if err != nil {
		return ServiceCatalog{}, err
	}
	var catalog ServiceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return ServiceCatalog{}, err
	}
	return catalog, nil
}

func writeServiceCatalog(path string, catalog ServiceCatalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func defaultServiceCatalog(layout Layout) ServiceCatalog {
	nodePath := bundledNodePath(layout)
	coreAppsDir := filepath.Join(layout.InstallDir, "core", "apps")
	adapterDir := filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter")

	return ServiceCatalog{
		Services: []ServiceSpec{
			{
				Name:            "api",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				HealthChecks: []HealthCheckSpec{
					{Name: "api-health", URL: "http://127.0.0.1:3000/health", TimeoutMS: 3000},
				},
				Commands: []CommandSpec{
					nodeCommand("api", nodePath, filepath.Join(coreAppsDir, "api"), filepath.Join(coreAppsDir, "api", "dist", "index.js")),
				},
			},
			{
				Name:            "device-service",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				HealthChecks: []HealthCheckSpec{
					{Name: "device-service-health", URL: "http://127.0.0.1:4010/health", TimeoutMS: 3000},
				},
				Commands: []CommandSpec{
					nodeCommand("api", nodePath, filepath.Join(coreAppsDir, "device-service"), filepath.Join(coreAppsDir, "device-service", "dist", "api", "main.js")),
					nodeCommand("outbox", nodePath, filepath.Join(coreAppsDir, "device-service"), filepath.Join(coreAppsDir, "device-service", "dist", "outbox", "main.js")),
				},
			},
			{
				Name:            "bot",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				Commands: []CommandSpec{
					nodeCommand("bot", nodePath, filepath.Join(coreAppsDir, "bot"), filepath.Join(coreAppsDir, "bot", "dist", "main.js")),
				},
			},
			{
				Name:            "worker",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				Commands: []CommandSpec{
					nodeCommand("preprocess", nodePath, filepath.Join(coreAppsDir, "worker"), filepath.Join(coreAppsDir, "worker", "dist", "main.js")),
					nodeCommand("access-events", nodePath, filepath.Join(coreAppsDir, "worker"), filepath.Join(coreAppsDir, "worker", "dist", "accessEvents", "main.js")),
					nodeCommand("outbox", nodePath, filepath.Join(coreAppsDir, "worker"), filepath.Join(coreAppsDir, "worker", "dist", "outbox", "main.js")),
					nodeCommand("retention", nodePath, filepath.Join(coreAppsDir, "worker"), filepath.Join(coreAppsDir, "worker", "dist", "retention", "main.js")),
					nodeCommand("monitoring", nodePath, filepath.Join(coreAppsDir, "worker"), filepath.Join(coreAppsDir, "worker", "dist", "monitoring", "main.js")),
				},
			},
			{
				Name:            "admin-ui",
				Kind:            "static-assets",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				StaticDir:       filepath.Join(coreAppsDir, "admin-ui", "dist"),
			},
			{
				Name:            "dahua-terminal-adapter",
				Kind:            "process-group",
				RequiresLicense: false,
				Env:             adapterServiceEnv(layout),
				HealthChecks: []HealthCheckSpec{
					{Name: "adapter-monitor", URL: "http://127.0.0.1:8091/monitor/snapshot", TimeoutMS: 3000},
				},
				Commands: []CommandSpec{
					nodeCommand("adapter", nodePath, adapterDir, filepath.Join(adapterDir, "dist", "src", "index.js")),
				},
			},
		},
	}
}

func bundledNodePath(layout Layout) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(layout.InstallDir, "runtime", "node", "node.exe")
	}
	return filepath.Join(layout.InstallDir, "runtime", "node", "bin", "node")
}

func nodeCommand(name, nodePath, workingDir, scriptPath string) CommandSpec {
	return CommandSpec{
		Name:       name,
		Executable: nodePath,
		Args:       []string{scriptPath},
		WorkingDir: workingDir,
	}
}

func commonServiceEnv(layout Layout) map[string]string {
	return map[string]string{
		"NODE_ENV":        "production",
		"SG_INSTALL_DIR":  layout.InstallDir,
		"SG_CONFIG_DIR":   layout.ConfigDir,
		"SG_DATA_DIR":     layout.DataDir,
		"SG_LOGS_DIR":     layout.LogsDir,
		"SG_LICENSES_DIR": layout.LicensesDir,
		"SG_RUNTIME_DIR":  layout.RuntimeDir,
		"SG_UPDATES_DIR":  layout.UpdatesDir,
		"SG_SUPERVISOR":   layout.Root,
		"SG_SERVICE_MODE": "supervisor",
	}
}

func adapterServiceEnv(layout Layout) map[string]string {
	env := commonServiceEnv(layout)
	env["LOG_PRETTY"] = "false"
	env["LOG_OUTPUT"] = "file"
	env["LOG_DIR"] = filepath.Join(layout.LogsDir, "dahua-terminal-adapter")
	env["SQLITE_PATH"] = filepath.Join(layout.DataDir, "dahua-terminal-adapter", "dahua-adapter.db")
	return env
}
