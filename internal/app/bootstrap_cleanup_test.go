package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/bootstrap"
)

func TestCleanupBootstrapWorkspaceRemovesSourceAndDeploy(t *testing.T) {
	supervisor := newTestApp(t)
	if err := supervisor.EnsureBootstrap(context.Background()); err != nil {
		t.Fatalf("ensure bootstrap: %v", err)
	}

	sourceRoot := filepath.Join(supervisor.bootstrap.Dir(), "source")
	deployRoot := filepath.Join(supervisor.bootstrap.Dir(), "deploy")
	keepRoot := filepath.Join(supervisor.layout.InstallDir, "core")

	for _, dir := range []string{sourceRoot, deployRoot, keepRoot} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	status := bootstrap.Status{
		Steps:      []bootstrap.StepStatus{{Name: "cleanup-workspace", State: "pending"}},
		SourceRoot: sourceRoot,
	}
	if err := supervisor.cleanupBootstrapWorkspace(&status); err != nil {
		t.Fatalf("cleanup bootstrap workspace: %v", err)
	}

	for _, path := range []string{sourceRoot, deployRoot} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, stat err=%v", path, err)
		}
	}
	if _, err := os.Stat(keepRoot); err != nil {
		t.Fatalf("expected install path to remain: %v", err)
	}
	if status.Steps[0].State != "succeeded" {
		t.Fatalf("expected cleanup step to succeed, got %s", status.Steps[0].State)
	}
}
