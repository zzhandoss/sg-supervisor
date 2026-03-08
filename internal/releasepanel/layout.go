package releasepanel

import (
	"os"
	"path/filepath"
)

type Layout struct {
	Root          string
	ConfigDir     string
	CacheDir      string
	JobsDir       string
	WorkspacesDir string
	LicensesDir   string
	ReleasesDir   string
	StatePath     string
	OwnerPath     string
}

func NewLayout(root string) Layout {
	return Layout{
		Root:          root,
		ConfigDir:     filepath.Join(root, "config"),
		CacheDir:      filepath.Join(root, "cache"),
		JobsDir:       filepath.Join(root, "jobs"),
		WorkspacesDir: filepath.Join(root, "workspaces"),
		LicensesDir:   filepath.Join(root, "licenses"),
		ReleasesDir:   filepath.Join(root, "releases"),
		StatePath:     filepath.Join(root, "config", "release-panel.json"),
		OwnerPath:     filepath.Join(root, "config", "owner.json"),
	}
}

func EnsureLayout(layout Layout) error {
	dirs := []string{
		layout.Root,
		layout.ConfigDir,
		layout.CacheDir,
		layout.JobsDir,
		layout.WorkspacesDir,
		layout.LicensesDir,
		layout.ReleasesDir,
		filepath.Join(layout.LicensesDir, "issued"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
