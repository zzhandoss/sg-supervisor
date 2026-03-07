package distribution

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sg-supervisor/internal/packaging"
)

func buildWindows(root string, stage packaging.AssembleReport) (Report, error) {
	outputDir := filepath.Join(root, "dist", "windows")
	if err := os.RemoveAll(outputDir); err != nil {
		return Report{}, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Report{}, err
	}

	wxsPath := filepath.Join(outputDir, "Product.wxs")
	scriptPath := filepath.Join(outputDir, "build-msi.ps1")
	if err := os.WriteFile(wxsPath, []byte(renderWiXSource(stage.OutputDir)), 0o644); err != nil {
		return Report{}, err
	}
	if err := os.WriteFile(scriptPath, []byte(renderWiXBuildScript()), 0o644); err != nil {
		return Report{}, err
	}

	report := Report{
		Platform:       "windows",
		StageDir:       stage.OutputDir,
		OutputDir:      outputDir,
		GeneratedFiles: []string{wxsPath, scriptPath},
	}
	candlePath, candleErr := exec.LookPath("candle.exe")
	lightPath, lightErr := exec.LookPath("light.exe")
	if candleErr != nil || lightErr != nil {
		report.Warnings = append(report.Warnings, "WiX toolchain was not found locally; generated Product.wxs and build-msi.ps1 only")
		return report, nil
	}
	artifactPath := filepath.Join(outputDir, "school-gate-windows-x64.msi")
	command := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "-CandlePath", candlePath, "-LightPath", lightPath, "-OutputPath", artifactPath)
	command.Dir = outputDir
	output, err := command.CombinedOutput()
	if err != nil {
		report.Warnings = append(report.Warnings, "MSI build failed: "+strings.TrimSpace(string(output)))
		return report, nil
	}
	report.ArtifactPath = artifactPath
	report.GeneratedFiles = append(report.GeneratedFiles, artifactPath)
	return report, nil
}

func renderWiXBuildScript() string {
	return strings.Join([]string{
		"param(",
		"  [Parameter(Mandatory=$true)][string]$CandlePath,",
		"  [Parameter(Mandatory=$true)][string]$LightPath,",
		"  [Parameter(Mandatory=$true)][string]$OutputPath",
		")",
		"$ErrorActionPreference = 'Stop'",
		"$root = Split-Path -Parent $MyInvocation.MyCommand.Path",
		"$wxs = Join-Path $root 'Product.wxs'",
		"$wixobj = Join-Path $root 'Product.wixobj'",
		"& $CandlePath -nologo -arch x64 $wxs -out $wixobj",
		"& $LightPath -nologo $wixobj -o $OutputPath",
		"Write-Host $OutputPath",
		"",
	}, "\n")
}
