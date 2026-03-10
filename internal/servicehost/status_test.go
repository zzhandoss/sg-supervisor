package servicehost

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"sg-supervisor/internal/config"
)

func TestQueryStatusReportsMissingConfig(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only service status")
	}

	root := t.TempDir()
	layout := config.NewLayout(root)
	cfg := config.SupervisorConfig{
		ProductName:   "School Gate",
		ListenAddress: "127.0.0.1:8787",
	}
	plan := BuildPlan(layout, cfg, filepath.Join(root, "sg-supervisor.exe"))
	if err := os.WriteFile(plan.WindowsWrapperPath, []byte("winsw"), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}

	status, err := QueryStatus(context.Background(), plan)
	if err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status.State != "config_missing" {
		t.Fatalf("expected config_missing, got %+v", status)
	}
	if status.LastError == "" {
		t.Fatalf("expected config missing error message")
	}
}
