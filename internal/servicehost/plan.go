package servicehost

import (
	"path/filepath"

	"sg-supervisor/internal/config"
)

type Plan struct {
	ServiceName          string   `json:"serviceName"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description"`
	BinaryPath           string   `json:"binaryPath"`
	Arguments            []string `json:"arguments"`
	WindowsWrapperPath   string   `json:"windowsWrapperPath"`
	WindowsConfigPath    string   `json:"windowsConfigPath"`
	LinuxUnitPath        string   `json:"linuxUnitPath"`
	WindowsInstallPath   string   `json:"windowsInstallPath"`
	WindowsUninstallPath string   `json:"windowsUninstallPath"`
	WindowsStartPath     string   `json:"windowsStartPath"`
	WindowsStopPath      string   `json:"windowsStopPath"`
	WorkingDirectory     string   `json:"workingDirectory"`
	ListenAddress        string   `json:"listenAddress"`
}

type RenderedArtifacts struct {
	Plan           Plan     `json:"plan"`
	WrittenFiles   []string `json:"writtenFiles"`
	InstallHints   []string `json:"installHints"`
	UninstallHints []string `json:"uninstallHints"`
}

func BuildPlan(layout config.Layout, cfg config.SupervisorConfig, binaryPath string) Plan {
	baseDir := filepath.Join(layout.RuntimeDir, "service-host")
	wrapperBase := filepath.Join(layout.Root, "school-gate-supervisor-service")
	return Plan{
		ServiceName:          "school-gate-supervisor",
		DisplayName:          "School Gate Supervisor",
		Description:          "School Gate installer, runtime supervisor, and control center host",
		BinaryPath:           binaryPath,
		Arguments:            []string{"serve", "--root", ".", "--listen", cfg.ListenAddress},
		WindowsWrapperPath:   wrapperBase + ".exe",
		WindowsConfigPath:    wrapperBase + ".xml",
		LinuxUnitPath:        filepath.Join(baseDir, "linux", "school-gate-supervisor.service"),
		WindowsInstallPath:   filepath.Join(baseDir, "windows", "install-service.ps1"),
		WindowsUninstallPath: filepath.Join(baseDir, "windows", "uninstall-service.ps1"),
		WindowsStartPath:     filepath.Join(baseDir, "windows", "start-service.ps1"),
		WindowsStopPath:      filepath.Join(baseDir, "windows", "stop-service.ps1"),
		WorkingDirectory:     layout.Root,
		ListenAddress:        cfg.ListenAddress,
	}
}
