package distribution

import (
	"errors"
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
	wixPath, err := findWiXExecutable()
	if err != nil {
		report.Warnings = append(report.Warnings, "WiX 6 CLI was not found locally; generated Product.wxs and build-msi.ps1 only")
		return report, nil
	}
	buildStageDir, cleanup, err := prepareWindowsBuildStage(stage.OutputDir)
	if err != nil {
		return Report{}, err
	}
	defer cleanup()
	mountedStageDir, unmount, err := mountWindowsBuildStage(buildStageDir)
	if err != nil {
		return Report{}, err
	}
	defer unmount()
	buildWXSPath := filepath.Join(outputDir, "Product.build.wxs")
	if err := os.WriteFile(buildWXSPath, []byte(renderWiXSource(mountedStageDir)), 0o644); err != nil {
		return Report{}, err
	}
	artifactPath := filepath.Join(outputDir, "school-gate-windows-x64.msi")
	absoluteWXSPath, err := filepath.Abs(buildWXSPath)
	if err != nil {
		return Report{}, err
	}
	absoluteArtifactPath, err := filepath.Abs(artifactPath)
	if err != nil {
		return Report{}, err
	}
	command := exec.Command(wixPath, "build", absoluteWXSPath, "-o", absoluteArtifactPath)
	command.Dir = outputDir
	output, err := command.CombinedOutput()
	if err != nil {
		report.Warnings = append(report.Warnings, "MSI build failed: "+strings.TrimSpace(string(output)))
		return report, nil
	}
	report.ArtifactPath = absoluteArtifactPath
	report.GeneratedFiles = append(report.GeneratedFiles, buildWXSPath, absoluteArtifactPath)
	return report, nil
}

func renderWiXBuildScript() string {
	return strings.Join([]string{
		"param(",
		"  [Parameter(Mandatory=$true)][string]$WixPath,",
		"  [Parameter(Mandatory=$true)][string]$OutputPath",
		")",
		"$ErrorActionPreference = 'Stop'",
		"$root = Split-Path -Parent $MyInvocation.MyCommand.Path",
		"$wxs = Join-Path $root 'Product.wxs'",
		"& $WixPath build $wxs -o $OutputPath",
		"Write-Host $OutputPath",
		"",
	}, "\n")
}

func findWiXExecutable() (string, error) {
	if path, err := exec.LookPath("wix.exe"); err == nil {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, ".dotnet", "tools", "wix.exe")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", errors.New("wix.exe not found")
}
