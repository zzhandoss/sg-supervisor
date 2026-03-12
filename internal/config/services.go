package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	StartupDelayMS  int               `json:"startupDelayMs,omitempty"`
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
	catalog, changed, err := loadAndReconcileServiceCatalog(path, layout)
	if err != nil {
		return err
	}
	if changed {
		return writeServiceCatalog(path, catalog)
	}
	return nil
}

func LoadServiceCatalog(layout Layout) (ServiceCatalog, error) {
	if err := EnsureServiceCatalog(layout); err != nil {
		return ServiceCatalog{}, err
	}
	catalog, _, err := loadAndReconcileServiceCatalog(ServicesFile(layout), layout)
	return catalog, err
}

func writeServiceCatalog(path string, catalog ServiceCatalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func loadAndReconcileServiceCatalog(path string, layout Layout) (ServiceCatalog, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServiceCatalog{}, false, err
	}
	var catalog ServiceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return ServiceCatalog{}, false, err
	}
	reconciled, changed := reconcileServiceCatalog(catalog, defaultServiceCatalog(layout))
	return reconciled, changed, nil
}

func reconcileServiceCatalog(current, defaults ServiceCatalog) (ServiceCatalog, bool) {
	defaultByName := make(map[string]ServiceSpec, len(defaults.Services))
	for _, service := range defaults.Services {
		defaultByName[service.Name] = service
	}

	changed := false
	reconciled := ServiceCatalog{Services: make([]ServiceSpec, 0, len(defaults.Services))}
	seen := make(map[string]struct{}, len(current.Services))
	for _, service := range current.Services {
		seen[service.Name] = struct{}{}
		def, ok := defaultByName[service.Name]
		if !ok {
			reconciled.Services = append(reconciled.Services, service)
			continue
		}
		next := service
		if requiresServiceReset(service, def) {
			next = def
			changed = true
		}
		reconciled.Services = append(reconciled.Services, next)
	}
	for _, def := range defaults.Services {
		if _, ok := seen[def.Name]; ok {
			continue
		}
		reconciled.Services = append(reconciled.Services, def)
		changed = true
	}
	return reconciled, changed
}

func requiresServiceReset(current, def ServiceSpec) bool {
	if current.Name == "admin-ui" && current.Kind != def.Kind {
		return true
	}
	if isSupervisorManagedService(current.Name) && len(current.Commands) == 0 && len(def.Commands) > 0 {
		return true
	}
	if isSupervisorManagedService(current.Name) && !reflect.DeepEqual(current.Commands, def.Commands) {
		return true
	}
	if isSupervisorManagedService(current.Name) && current.StartupDelayMS != def.StartupDelayMS {
		return true
	}
	return false
}

func defaultServiceCatalog(layout Layout) ServiceCatalog {
	nodePath := bundledNodePath(layout)
	coreRootDir := filepath.Join(layout.InstallDir, "core")
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
					nodeCommand("api", nodePath, coreRootDir, filepath.Join(coreAppsDir, "api", "dist", "index.js")),
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
					nodeCommand("api", nodePath, coreRootDir, filepath.Join(coreAppsDir, "device-service", "dist", "api", "main.js")),
					nodeCommand("outbox", nodePath, coreRootDir, filepath.Join(coreAppsDir, "device-service", "dist", "outbox", "main.js")),
				},
			},
			{
				Name:            "bot",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				Commands: []CommandSpec{
					nodeCommand("bot", nodePath, coreRootDir, filepath.Join(coreAppsDir, "bot", "dist", "main.js")),
				},
			},
			{
				Name:            "worker",
				Kind:            "process-group",
				RequiresLicense: true,
				StartupDelayMS:  750,
				Env:             commonServiceEnv(layout),
				Commands: []CommandSpec{
					nodeCommand("preprocess", nodePath, coreRootDir, filepath.Join(coreAppsDir, "worker", "dist", "main.js")),
					nodeCommand("access-events", nodePath, coreRootDir, filepath.Join(coreAppsDir, "worker", "dist", "accessEvents", "main.js")),
					nodeCommand("outbox", nodePath, coreRootDir, filepath.Join(coreAppsDir, "worker", "dist", "outbox", "main.js")),
					nodeCommand("retention", nodePath, coreRootDir, filepath.Join(coreAppsDir, "worker", "dist", "retention", "main.js")),
					nodeCommand("monitoring", nodePath, coreRootDir, filepath.Join(coreAppsDir, "worker", "dist", "monitoring", "main.js")),
				},
			},
			{
				Name:            "admin-ui",
				Kind:            "process-group",
				RequiresLicense: true,
				Env:             commonServiceEnv(layout),
				HealthChecks: []HealthCheckSpec{
					{Name: "admin-ui-health", URL: "http://127.0.0.1:5000/healthz", TimeoutMS: 3000},
				},
				Commands: []CommandSpec{
					nodeCommand("admin-ui", nodePath, coreRootDir, filepath.Join(coreAppsDir, "admin-ui", ".output", "server", "index.mjs")),
				},
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
		"NODE_ENV":              "production",
		"LOG_LEVEL":             "info",
		"LOG_DIR":               filepath.Join(layout.LogsDir, "school-gate"),
		"LOG_MAX_BYTES":         "104857600",
		"LOG_RETENTION_DAYS":    "7",
		"BACKUP_DIR":            filepath.Join(layout.BackupsDir, "school-gate"),
		"BACKUP_LICENSE_DIR":    layout.LicensesDir,
		"BACKUP_INCLUDE_LOGS":   "false",
		"BACKUP_LOGS_MAX_FILES": "4",
		"BACKUP_KEEP_NIGHTLY":   "14",
		"BACKUP_KEEP_PREUPDATE": "3",
		"SG_INSTALL_DIR":        layout.InstallDir,
		"SG_CONFIG_DIR":         layout.ConfigDir,
		"SG_DATA_DIR":           layout.DataDir,
		"SG_LOGS_DIR":           layout.LogsDir,
		"SG_LICENSES_DIR":       layout.LicensesDir,
		"SG_RUNTIME_DIR":        layout.RuntimeDir,
		"SG_UPDATES_DIR":        layout.UpdatesDir,
		"SG_SUPERVISOR":         layout.Root,
		"SG_SERVICE_MODE":       "supervisor",
	}
}

func adapterServiceEnv(layout Layout) map[string]string {
	env := commonServiceEnv(layout)
	env["LOG_PRETTY"] = "false"
	env["LOG_OUTPUT"] = "file"
	env["LOG_DIR"] = filepath.Join(layout.LogsDir, "dahua-terminal-adapter")
	env["SQLITE_PATH"] = filepath.Join(layout.DataDir, "dahua-terminal-adapter", "dahua-adapter.db")
	env["BACKUP_DIR"] = filepath.Join(layout.BackupsDir, "dahua-terminal-adapter")
	env["BACKUP_LICENSE_DIR"] = layout.LicensesDir
	env["BACKUP_INCLUDE_LOGS"] = "true"
	return env
}

func cloneEnv(source map[string]string) map[string]string {
	if len(source) == 0 {
		return map[string]string{}
	}
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

func isSupervisorManagedService(name string) bool {
	switch name {
	case "api", "device-service", "bot", "worker", "admin-ui", "dahua-terminal-adapter":
		return true
	default:
		return false
	}
}
