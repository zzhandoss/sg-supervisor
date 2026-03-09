package runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
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
	env := append([]string(nil), os.Environ()...)
	keys := make([]string, 0, len(extra))
	for key := range extra {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+extra[key])
	}
	return env
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
