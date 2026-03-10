package servicehost

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Render(plan Plan) (RenderedArtifacts, error) {
	files := map[string]string{
		plan.WindowsConfigPath:    renderWindowsServiceConfig(plan),
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
		"$wrapperPath = '" + escapePowerShell(plan.WindowsWrapperPath) + "'",
		"if (-not (Test-Path $wrapperPath)) { throw 'WinSW wrapper is missing: ' + $wrapperPath }",
		"& $wrapperPath install",
		"sc.exe description $serviceName '" + escapePowerShell(plan.Description) + "'",
		"sc.exe config $serviceName start= demand",
		"",
	}, "\n")
}

func renderWindowsUninstallScript(plan Plan) string {
	return strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"$wrapperPath = '" + escapePowerShell(plan.WindowsWrapperPath) + "'",
		"if (Test-Path $wrapperPath) {",
		"  try { & $wrapperPath stop } catch {}",
		"  & $wrapperPath uninstall",
		"}",
		"",
	}, "\n")
}

func renderWindowsStartScript(plan Plan) string {
	return "$ErrorActionPreference = 'Stop'\n& '" + escapePowerShell(plan.WindowsWrapperPath) + "' start\n"
}

func renderWindowsStopScript(plan Plan) string {
	return "$ErrorActionPreference = 'Stop'\n& '" + escapePowerShell(plan.WindowsWrapperPath) + "' stop\n"
}

func renderWindowsServiceConfig(plan Plan) string {
	return strings.Join([]string{
		"<service>",
		"  <id>" + plan.ServiceName + "</id>",
		"  <name>" + escapeXML(plan.DisplayName) + "</name>",
		"  <description>" + escapeXML(plan.Description) + "</description>",
		"  <executable>%BASE%\\" + escapeXML(filepath.Base(plan.BinaryPath)) + "</executable>",
		"  <arguments>" + escapeXML(joinWindowsServiceArgs(plan.Arguments)) + "</arguments>",
		"  <workingdirectory>%BASE%</workingdirectory>",
		"</service>",
		"",
	}, "\n")
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

func joinWindowsServiceArgs(args []string) string {
	escaped := make([]string, 0, len(args))
	for _, arg := range args {
		escaped = append(escaped, `"`+strings.ReplaceAll(arg, `"`, `&quot;`)+`"`)
	}
	return strings.Join(escaped, " ")
}

func escapeXML(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	value = strings.ReplaceAll(value, `"`, "&quot;")
	return strings.ReplaceAll(value, "'", "&apos;")
}
