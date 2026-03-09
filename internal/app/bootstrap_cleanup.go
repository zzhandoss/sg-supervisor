package app

import (
	"os"
	"path/filepath"

	"sg-supervisor/internal/bootstrap"
)

func (a *App) cleanupBootstrapWorkspace(status *bootstrap.Status) error {
	if err := a.markBootstrapStep(status, "cleanup-workspace", "running", "Cleaning bootstrap workspace"); err != nil {
		return err
	}
	paths := []string{
		status.SourceRoot,
		filepath.Join(a.bootstrap.Dir(), "deploy"),
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			a.failBootstrap(*status, "cleanup-workspace", err)
			return err
		}
	}
	return a.completeBootstrapStep(status, "cleanup-workspace", "Bootstrap workspace cleaned")
}
