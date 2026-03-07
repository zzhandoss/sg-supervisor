package servicehost

import (
	"errors"
	"path/filepath"
	"runtime"
)

type Action struct {
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	Args          []string `json:"args,omitempty"`
	IgnoreFailure bool     `json:"ignoreFailure,omitempty"`
}

func InstallActions(plan Plan) ([]Action, error) {
	return InstallActionsForTarget(plan, runtime.GOOS)
}

func RepairActions(plan Plan) ([]Action, error) {
	return InstallActionsForTarget(plan, runtime.GOOS)
}

func UninstallActions(plan Plan) ([]Action, error) {
	return UninstallActionsForTarget(plan, runtime.GOOS)
}

func InstallActionsForTarget(plan Plan, targetOS string) ([]Action, error) {
	return installActionsForOS(plan, targetOS)
}

func UninstallActionsForTarget(plan Plan, targetOS string) ([]Action, error) {
	return uninstallActionsForOS(plan, targetOS)
}

func installActionsForOS(plan Plan, targetOS string) ([]Action, error) {
	switch targetOS {
	case "windows":
		return []Action{
			powerShellAction("install-service", plan.WindowsInstallPath),
		}, nil
	case "linux":
		return []Action{
			{Name: "install-unit", Command: "cp", Args: []string{plan.LinuxUnitPath, linuxUnitInstallPath(plan)}},
			{Name: "systemd-daemon-reload", Command: "systemctl", Args: []string{"daemon-reload"}},
			{Name: "enable-service", Command: "systemctl", Args: []string{"enable", "--now", plan.ServiceName}},
		}, nil
	default:
		return nil, errors.New("unsupported service-host platform")
	}
}

func uninstallActionsForOS(plan Plan, targetOS string) ([]Action, error) {
	switch targetOS {
	case "windows":
		return []Action{
			powerShellAction("uninstall-service", plan.WindowsUninstallPath),
		}, nil
	case "linux":
		return []Action{
			{Name: "disable-service", Command: "systemctl", Args: []string{"disable", "--now", plan.ServiceName}, IgnoreFailure: true},
			{Name: "remove-unit", Command: "rm", Args: []string{"-f", linuxUnitInstallPath(plan)}},
			{Name: "systemd-daemon-reload", Command: "systemctl", Args: []string{"daemon-reload"}},
		}, nil
	default:
		return nil, errors.New("unsupported service-host platform")
	}
}

func powerShellAction(name, scriptPath string) Action {
	return Action{
		Name:    name,
		Command: "powershell.exe",
		Args:    []string{"-ExecutionPolicy", "Bypass", "-File", scriptPath},
	}
}

func linuxUnitInstallPath(plan Plan) string {
	return filepath.Join("/etc/systemd/system", plan.ServiceName+".service")
}
