package servicehost

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Status struct {
	Supported   bool   `json:"supported"`
	ServiceName string `json:"serviceName"`
	Installed   bool   `json:"installed"`
	State       string `json:"state"`
	StartMode   string `json:"startMode"`
	WrapperPath string `json:"wrapperPath,omitempty"`
	ConfigPath  string `json:"configPath,omitempty"`
	LastError   string `json:"lastError,omitempty"`
	Description string `json:"description,omitempty"`
}

func QueryStatus(ctx context.Context, plan Plan) (Status, error) {
	status := Status{
		Supported:   runtime.GOOS == "windows",
		ServiceName: plan.ServiceName,
		WrapperPath: plan.WindowsWrapperPath,
		ConfigPath:  plan.WindowsConfigPath,
		Description: plan.Description,
		State:       "unsupported",
		StartMode:   "unknown",
	}
	if runtime.GOOS != "windows" {
		return status, nil
	}
	if _, err := os.Stat(plan.WindowsWrapperPath); err != nil {
		status.State = "wrapper_missing"
		status.LastError = err.Error()
		return status, nil
	}
	if _, err := os.Stat(plan.WindowsConfigPath); err != nil {
		status.State = "config_missing"
		status.LastError = err.Error()
		return status, nil
	}

	output, err := exec.CommandContext(ctx, plan.WindowsWrapperPath, "status").CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil && text == "" {
		return status, err
	}

	switch text {
	case "NonExistent":
		status.State = "not_installed"
		return status, nil
	case "Started":
		status.Installed = true
		status.State = "running"
	case "Stopped":
		status.Installed = true
		status.State = "stopped"
	default:
		status.State = "error"
		status.LastError = text
	}

	startMode, err := queryStartMode(ctx, plan.ServiceName)
	if err != nil {
		status.LastError = err.Error()
		return status, nil
	}
	status.StartMode = startMode
	return status, nil
}

func InstallService(ctx context.Context, plan Plan) error {
	if runtime.GOOS != "windows" {
		return errors.New("windows service management is only available on Windows")
	}
	if _, err := os.Stat(plan.WindowsWrapperPath); err != nil {
		return err
	}
	if _, err := exec.CommandContext(ctx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-File", plan.WindowsInstallPath).CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func StartService(ctx context.Context, plan Plan) error {
	return runPowerShellScript(ctx, plan.WindowsStartPath)
}

func StopService(ctx context.Context, plan Plan) error {
	return runPowerShellScript(ctx, plan.WindowsStopPath)
}

func RemoveService(ctx context.Context, plan Plan) error {
	return runPowerShellScript(ctx, plan.WindowsUninstallPath)
}

func EnableAutostart(ctx context.Context, plan Plan) error {
	return setStartMode(ctx, plan.ServiceName, "auto")
}

func DisableAutostart(ctx context.Context, plan Plan) error {
	return setStartMode(ctx, plan.ServiceName, "demand")
}

func ScheduleStart(ctx context.Context, plan Plan, delay time.Duration) error {
	if runtime.GOOS != "windows" {
		return errors.New("windows service management is only available on Windows")
	}
	seconds := int(delay.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	script := "Start-Sleep -Seconds " + strconv.Itoa(seconds) + "; & '" + strings.ReplaceAll(plan.WindowsWrapperPath, "'", "''") + "' start"
	command := exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
	return command.Start()
}

func runPowerShellScript(ctx context.Context, scriptPath string) error {
	if runtime.GOOS != "windows" {
		return errors.New("windows service management is only available on Windows")
	}
	command := exec.CommandContext(ctx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return err
	}
	return errors.New(message)
}

func setStartMode(ctx context.Context, serviceName, mode string) error {
	command := exec.CommandContext(ctx, "sc.exe", "config", serviceName, "start=", mode)
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return err
	}
	return errors.New(message)
}

func queryStartMode(ctx context.Context, serviceName string) (string, error) {
	command := exec.CommandContext(ctx, "sc.exe", "qc", serviceName)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return "", err
		}
		return "", errors.New(message)
	}
	body := string(output)
	switch {
	case strings.Contains(body, "AUTO_START"):
		return "automatic", nil
	case strings.Contains(body, "DEMAND_START"):
		return "manual", nil
	case strings.Contains(body, "DISABLED"):
		return "disabled", nil
	default:
		return "unknown", nil
	}
}
