package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultServiceCatalogAdminUIUsesNodeProcess(t *testing.T) {
	layout := NewLayout(filepath.Join("C:", "test-root"))

	catalog := defaultServiceCatalog(layout)
	var adminUI ServiceSpec
	for _, service := range catalog.Services {
		if service.Name == "admin-ui" {
			adminUI = service
			break
		}
	}

	if adminUI.Kind != "process-group" {
		t.Fatalf("expected admin-ui to be a process-group, got %q", adminUI.Kind)
	}
	if len(adminUI.Commands) != 1 {
		t.Fatalf("expected one admin-ui command, got %d", len(adminUI.Commands))
	}
	if len(adminUI.HealthChecks) != 1 || adminUI.HealthChecks[0].URL != "http://127.0.0.1:5000/healthz" {
		t.Fatalf("unexpected admin-ui health checks: %#v", adminUI.HealthChecks)
	}

	expectedBinary := filepath.Join(layout.InstallDir, "runtime", "node", "bin", "node")
	if runtime.GOOS == "windows" {
		expectedBinary = filepath.Join(layout.InstallDir, "runtime", "node", "node.exe")
	}
	if adminUI.Commands[0].Executable != expectedBinary {
		t.Fatalf("expected admin-ui binary %q, got %q", expectedBinary, adminUI.Commands[0].Executable)
	}
	expectedScript := filepath.Join(layout.InstallDir, "core", "apps", "admin-ui", ".output", "server", "index.mjs")
	if len(adminUI.Commands[0].Args) != 1 || adminUI.Commands[0].Args[0] != expectedScript {
		t.Fatalf("unexpected admin-ui args: %#v", adminUI.Commands[0].Args)
	}
}
