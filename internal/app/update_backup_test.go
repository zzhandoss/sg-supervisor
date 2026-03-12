package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreUpdateBackupCommandsIncludeSchoolGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	nodePath := nodeExecutablePath(app.root)
	coreScript := filepath.Join(app.layout.InstallDir, "core", "packages", "ops", "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(nodePath), 0o755); err != nil {
		t.Fatalf("mkdir node dir: %v", err)
	}
	if err := os.WriteFile(nodePath, []byte("node"), 0o644); err != nil {
		t.Fatalf("write node: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(coreScript), 0o755); err != nil {
		t.Fatalf("mkdir core script dir: %v", err)
	}
	if err := os.WriteFile(coreScript, []byte("ops"), 0o644); err != nil {
		t.Fatalf("write core script: %v", err)
	}
	commands := app.preUpdateBackupCommands()
	if len(commands) == 0 {
		t.Fatalf("expected backup commands")
	}
	core := commands[0]
	if core.Name != "school-gate" {
		t.Fatalf("expected first backup command to be school-gate, got %q", core.Name)
	}
	if core.Dir != filepath.Join(app.layout.InstallDir, "core") {
		t.Fatalf("unexpected school-gate backup dir: %q", core.Dir)
	}
}

func TestPreUpdateBackupCommandsIncludeAdapterWhenInstalled(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	nodePath := nodeExecutablePath(app.root)
	coreScript := filepath.Join(app.layout.InstallDir, "core", "packages", "ops", "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(nodePath), 0o755); err != nil {
		t.Fatalf("mkdir node dir: %v", err)
	}
	if err := os.WriteFile(nodePath, []byte("node"), 0o644); err != nil {
		t.Fatalf("write node: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(coreScript), 0o755); err != nil {
		t.Fatalf("mkdir core script dir: %v", err)
	}
	if err := os.WriteFile(coreScript, []byte("ops"), 0o644); err != nil {
		t.Fatalf("write core script: %v", err)
	}
	adapterScript := filepath.Join(app.layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src", "ops", "backup", "backup-cli.js")
	if err := os.MkdirAll(filepath.Dir(adapterScript), 0o755); err != nil {
		t.Fatalf("mkdir adapter script dir: %v", err)
	}
	if err := os.WriteFile(adapterScript, []byte("adapter backup"), 0o644); err != nil {
		t.Fatalf("write adapter script: %v", err)
	}

	commands := app.preUpdateBackupCommands()
	if len(commands) != 2 {
		t.Fatalf("expected 2 backup commands, got %d", len(commands))
	}
	adapter := commands[1]
	if adapter.Name != "dahua-terminal-adapter" {
		t.Fatalf("expected adapter backup command, got %q", adapter.Name)
	}
	if adapter.Dir != app.root {
		t.Fatalf("expected adapter backup dir to be supervisor root, got %q", adapter.Dir)
	}
	if want := filepath.Join(app.layout.BackupsDir, "dahua-terminal-adapter"); adapter.Args[7] != want {
		t.Fatalf("unexpected adapter backups dir arg: %q", adapter.Args[7])
	}
}

func TestWithTemporaryRootFilesRestoresOriginalFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceEnv := filepath.Join(root, "adapter.env")
	targetEnv := filepath.Join(root, ".env")
	if err := os.WriteFile(sourceEnv, []byte("A=1\n"), 0o644); err != nil {
		t.Fatalf("write source env: %v", err)
	}
	if err := os.WriteFile(targetEnv, []byte("ORIGINAL=1\n"), 0o644); err != nil {
		t.Fatalf("write target env: %v", err)
	}

	if _, err := withTemporaryRootFiles(map[string]string{targetEnv: sourceEnv}, func() (string, error) {
		data, err := os.ReadFile(targetEnv)
		if err != nil {
			return "", err
		}
		if string(data) != "A=1\n" {
			t.Fatalf("unexpected temporary env content: %q", string(data))
		}
		return "ok", nil
	}); err != nil {
		t.Fatalf("withTemporaryRootFiles: %v", err)
	}

	data, err := os.ReadFile(targetEnv)
	if err != nil {
		t.Fatalf("read restored env: %v", err)
	}
	if string(data) != "ORIGINAL=1\n" {
		t.Fatalf("unexpected restored env content: %q", string(data))
	}
}
