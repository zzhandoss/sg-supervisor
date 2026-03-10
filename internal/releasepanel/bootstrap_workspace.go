package releasepanel

import (
	"errors"
	"os"
	"path/filepath"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/servicehost"
)

func prepareBootstrapWorkspace(root, platform string, state State, assets WorkspaceAssets) error {
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		return err
	}
	if err := extractArchive(assets.NodePath, filepath.Join(layout.InstallDir, "runtime", "node")); err != nil {
		return err
	}
	if err := prepareNodeRuntime(filepath.Join(layout.InstallDir, "runtime", "node"), platform); err != nil {
		return err
	}
	if err := writeSupervisorConfig(layout.ConfigFile, state); err != nil {
		return err
	}
	if platform == "windows" {
		targetPath := filepath.Join(layout.Root, "school-gate-supervisor-service.exe")
		if err := copyArtifact(assets.WinSWPath, targetPath); err != nil {
			return err
		}
		if err := renderWindowsServiceArtifacts(layout); err != nil {
			return err
		}
	}
	return validateBootstrapWorkspace(layout, platform)
}

func stageBootstrapBundle(root, sourcePath string) (string, error) {
	layout := config.NewLayout(root)
	targetPath := filepath.Join(layout.InstallDir, "bootstrap", filepath.Base(sourcePath))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", err
	}
	if err := copyFile(sourcePath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

func validateBootstrapWorkspace(layout config.Layout, platform string) error {
	required := []string{
		layout.ConfigFile,
	}
	if platform == "windows" {
		required = append(required, filepath.Join(layout.InstallDir, "runtime", "node", "node.exe"))
		required = append(required, filepath.Join(layout.Root, "school-gate-supervisor-service.exe"))
		required = append(required, filepath.Join(layout.Root, "school-gate-supervisor-service.xml"))
	} else {
		required = append(required, filepath.Join(layout.InstallDir, "runtime", "node", "bin", "node"))
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			return errors.New("bootstrap workspace is missing required path: " + path)
		}
	}
	return nil
}

func renderWindowsServiceArtifacts(layout config.Layout) error {
	cfg, err := config.LoadOrCreate(layout.ConfigFile)
	if err != nil {
		return err
	}
	plan := servicehost.BuildPlan(layout, cfg, filepath.Join(layout.Root, "sg-supervisor.exe"))
	_, err = servicehost.Render(plan)
	return err
}
