package config

import (
	"os"
	"strings"
	"testing"
)

func TestInternalRuntimeStoreGeneratesAndPersistsSecrets(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewInternalRuntimeStore(layout)
	if err := store.Ensure(); err != nil {
		t.Fatalf("ensure internal runtime: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("load internal runtime: %v", err)
	}
	if cfg.CoreToken == "" || cfg.AdminJWTSecret == "" || cfg.DeviceServiceToken == "" {
		t.Fatalf("expected generated secrets, got %+v", cfg)
	}

	data, err := os.ReadFile(InternalRuntimeEnvFile(layout))
	if err != nil {
		t.Fatalf("read internal env file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "CORE_TOKEN="+cfg.CoreToken+"\n") {
		t.Fatalf("missing core token in env file: %q", content)
	}
	if !strings.Contains(content, "DEVICE_SERVICE_INTERNAL_TOKEN="+cfg.DeviceServiceInternalKey+"\n") {
		t.Fatalf("missing device service internal token in env file: %q", content)
	}
}
