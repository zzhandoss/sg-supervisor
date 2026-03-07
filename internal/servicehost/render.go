package servicehost

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Render(plan Plan) (RenderedArtifacts, error) {
	files := map[string]string{
		plan.LinuxUnitPath:        renderLinuxUnit(plan),
		plan.WindowsInstallPath:   renderWindowsInstallScript(plan),
		plan.WindowsUninstallPath: renderWindowsUninstallScript(plan),
		plan.WindowsStartPath:     renderWindowsStartScript(plan),
		plan.WindowsStopPath:      renderWindowsStopScript(plan),
	}

	written := make([]string, 0, len(files))
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return RenderedArtifacts{}, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return RenderedArtifacts{}, err
		}
		written = append(written, path)
	}

	return RenderedArtifacts{
		Plan:         plan,
		WrittenFiles: written,
		InstallHints: []string{
			fmt.Sprintf("Linux: sudo cp %q /etc/systemd/system/%s && sudo systemctl daemon-reload && sudo systemctl enable --now %s", plan.LinuxUnitPath, plan.ServiceName+".service", plan.ServiceName),
			fmt.Sprintf("Windows: powershell -ExecutionPolicy Bypass -File %q", plan.WindowsInstallPath),
		},
		UninstallHints: []string{
			fmt.Sprintf("Linux: sudo systemctl disable --now %s && sudo rm /etc/systemd/system/%s && sudo systemctl daemon-reload", plan.ServiceName, plan.ServiceName+".service"),
			fmt.Sprintf("Windows: powershell -ExecutionPolicy Bypass -File %q", plan.WindowsUninstallPath),
		},
	}, nil
}

func renderLinuxUnit(plan Plan) string {
	return strings.Join([]string{
		"[Unit]",
		"Description=" + plan.Description,
		"After=network.target",
		"",
		"[Service]",
		"Type=simple",
		"WorkingDirectory=" + plan.WorkingDirectory,
		"ExecStart=" + quoteLinux(plan.BinaryPath) + " " + joinLinuxArgs(plan.Arguments),
		"Restart=on-failure",
		"RestartSec=5",
		"Environment=SG_SERVICE_HOST=systemd",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
		"",
	}, "\n")
}

func renderWindowsInstallScript(plan Plan) string {
	return strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"$serviceName = '" + plan.ServiceName + "'",
		"$displayName = '" + plan.DisplayName + "'",
		"$binaryPath = '" + escapePowerShell(plan.BinaryPath) + "'",
		"$args = '" + escapePowerShell(joinWindowsArgs(plan.Arguments)) + "'",
		`$binPath = '"' + $binaryPath + '" ' + $args`,
		"$service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue",
		"if ($null -eq $service) {",
		"  sc.exe create $serviceName binPath= $binPath start= auto DisplayName= $displayName",
		"} else {",
		"  sc.exe config $serviceName binPath= $binPath start= auto",
		"}",
		"sc.exe description $serviceName '" + escapePowerShell(plan.Description) + "'",
		"if ((Get-Service -Name $serviceName).Status -ne 'Running') { sc.exe start $serviceName }",
		"",
	}, "\n")
}

func renderWindowsUninstallScript(plan Plan) string {
	return strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"$serviceName = '" + plan.ServiceName + "'",
		"$service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue",
		"if ($null -ne $service) {",
		"  if ($service.Status -ne 'Stopped') {",
		"    sc.exe stop $serviceName",
		"    Start-Sleep -Seconds 2",
		"  }",
		"  sc.exe delete $serviceName",
		"}",
		"",
	}, "\n")
}

func renderWindowsStartScript(plan Plan) string {
	return "$ErrorActionPreference = 'Stop'\nsc.exe start " + plan.ServiceName + "\n"
}

func renderWindowsStopScript(plan Plan) string {
	return "$ErrorActionPreference = 'Stop'\nsc.exe stop " + plan.ServiceName + "\n"
}

func quoteLinux(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func joinLinuxArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, quoteLinux(arg))
	}
	return strings.Join(quoted, " ")
}

func joinWindowsArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, `"`+strings.ReplaceAll(arg, `"`, `\"`)+`"`)
	}
	return strings.Join(quoted, " ")
}

func escapePowerShell(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
