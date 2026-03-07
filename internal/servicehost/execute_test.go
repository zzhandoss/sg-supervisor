package servicehost

import (
	"context"
	"errors"
	"testing"
)

type recordingRunner struct {
	actions []Action
	failAt  string
}

func (r *recordingRunner) Run(_ context.Context, action Action) error {
	r.actions = append(r.actions, action)
	if action.Name == r.failAt {
		return errors.New("boom")
	}
	return nil
}

func TestInstallActionsForWindows(t *testing.T) {
	actions, err := installActionsForOS(Plan{WindowsInstallPath: `C:\svc\install-service.ps1`}, "windows")
	if err != nil {
		t.Fatalf("install actions: %v", err)
	}
	if len(actions) != 1 || actions[0].Command != "powershell.exe" {
		t.Fatalf("unexpected actions: %+v", actions)
	}
}

func TestUninstallActionsForLinux(t *testing.T) {
	actions, err := uninstallActionsForOS(Plan{ServiceName: "school-gate-supervisor"}, "linux")
	if err != nil {
		t.Fatalf("uninstall actions: %v", err)
	}
	if len(actions) != 3 {
		t.Fatalf("expected 3 uninstall actions, got %d", len(actions))
	}
	if actions[0].IgnoreFailure {
		t.Fatalf("expected disable-service failure to surface")
	}
}

func TestExecuteActionsIgnoresMarkedFailures(t *testing.T) {
	runner := &recordingRunner{failAt: "disable-service"}
	actions := []Action{
		{Name: "disable-service", IgnoreFailure: true},
		{Name: "remove-unit"},
	}
	if err := executeActions(context.Background(), runner, actions); err != nil {
		t.Fatalf("expected ignore failure behavior: %v", err)
	}
	if len(runner.actions) != 2 {
		t.Fatalf("expected both actions to execute, got %d", len(runner.actions))
	}
}
