package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"sg-supervisor/internal/config"
)

func buildInitialStatus(service config.ServiceSpec) ServiceStatus {
	status := ServiceStatus{
		Name:            service.Name,
		Kind:            service.Kind,
		Configured:      config.ServiceConfigured(service),
		RequiresLicense: service.RequiresLicense,
		State:           "stopped",
		Readiness:       "not_ready",
		Reachability:    "unknown",
		StaticDir:       service.StaticDir,
		Components:      make([]ComponentStatus, 0, len(service.Commands)),
	}
	for _, command := range service.Commands {
		status.Components = append(status.Components, ComponentStatus{
			Name:       command.Name,
			Executable: command.Executable,
			State:      "stopped",
		})
	}
	return status
}

func aggregateServiceState(components []ComponentStatus, kind string) string {
	if kind == "static-assets" {
		return "ready"
	}
	if len(components) == 0 {
		return "stopped"
	}

	hasRunning := false
	for _, component := range components {
		switch component.State {
		case "error":
			return "error"
		case "running":
			hasRunning = true
		}
	}
	if hasRunning {
		return "running"
	}
	return "stopped"
}

func serviceLastError(components []ComponentStatus) string {
	for _, component := range components {
		if component.LastError != "" {
			return component.Name + ": " + component.LastError
		}
	}
	return ""
}

func mergeEnv(extra map[string]string) []string {
	merged := make(map[string]string)
	original := make(map[string]string)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		normalized := normalizeEnvKey(key)
		if _, exists := original[normalized]; !exists {
			original[normalized] = key
		}
		merged[normalized] = value
	}
	for key, value := range extra {
		normalized := normalizeEnvKey(key)
		original[normalized] = key
		merged[normalized] = value
	}
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, original[key]+"="+merged[key])
	}
	return env
}

func normalizeEnvKey(key string) string {
	if os.PathListSeparator == ';' {
		return strings.ToUpper(key)
	}
	return key
}

func exitError(err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return fmt.Sprintf("exit status %d", status.ExitStatus())
		}
	}
	return err.Error()
}
