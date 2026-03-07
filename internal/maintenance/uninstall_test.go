package maintenance

import (
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/config"
)

func TestExecuteUninstallKeepState(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	targetFile := filepath.Join(layout.InstallDir, "core.txt")
	if err := os.WriteFile(targetFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write install file: %v", err)
	}

	report, err := ExecuteUninstall(layout, ModeKeepState)
	if err != nil {
		t.Fatalf("execute uninstall: %v", err)
	}
	if report.Mode != ModeKeepState {
		t.Fatalf("unexpected mode: %s", report.Mode)
	}
	if _, err := os.Stat(layout.InstallDir); !os.IsNotExist(err) {
		t.Fatalf("expected install dir removed")
	}
	if _, err := os.Stat(layout.ConfigDir); err != nil {
		t.Fatalf("expected config dir kept: %v", err)
	}
}

func TestExecuteUninstallFullWipe(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	report, err := ExecuteUninstall(layout, ModeFullWipe)
	if err != nil {
		t.Fatalf("execute uninstall: %v", err)
	}
	if report.Mode != ModeFullWipe {
		t.Fatalf("unexpected mode: %s", report.Mode)
	}
	if _, err := os.Stat(layout.ConfigDir); !os.IsNotExist(err) {
		t.Fatalf("expected config dir removed")
	}
	if _, err := os.Stat(layout.DataDir); !os.IsNotExist(err) {
		t.Fatalf("expected data dir removed")
	}
}
