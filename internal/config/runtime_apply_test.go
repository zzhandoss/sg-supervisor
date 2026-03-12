package config

import (
	"path/filepath"
	"testing"
)

func TestApplyRuntimeConfigInjectsCoreAndAdapterEnv(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{Name: "api", Kind: "process-group", Env: map[string]string{"NODE_ENV": "production"}},
			{Name: "bot", Kind: "process-group", Env: map[string]string{}},
			{Name: "dahua-terminal-adapter", Kind: "process-group", Env: map[string]string{}},
		},
	}
	internal := InternalRuntimeConfig{
		CoreToken:                "core-token",
		CoreHMACSecret:           "core-hmac",
		AdminJWTSecret:           "admin-jwt-secret-admin-jwt-secret",
		DeviceServiceToken:       "device-token",
		DeviceServiceInternalKey: "device-internal",
		BotInternalToken:         "bot-internal",
	}

	applied := ApplyRuntimeConfig(layout, catalog, ProductConfig{TelegramBotToken: "bot-token"}, internal)

	apiEnv := applied.Services[0].Env
	if apiEnv["CORE_TOKEN"] != "core-token" || apiEnv["DB_FILE"] != filepath.Join(layout.DataDir, "school-gate", "app.db") {
		t.Fatalf("unexpected api env: %+v", apiEnv)
	}
	if apiEnv["LOG_DIR"] != filepath.Join(layout.LogsDir, "school-gate") {
		t.Fatalf("expected canonical log dir, got %+v", apiEnv)
	}
	if apiEnv["BACKUP_DIR"] != filepath.Join(layout.BackupsDir, "school-gate") {
		t.Fatalf("expected canonical backup dir, got %+v", apiEnv)
	}

	botEnv := applied.Services[1].Env
	if botEnv["BOT_INTERNAL_TOKEN"] != "bot-internal" || botEnv["TELEGRAM_BOT_TOKEN"] != "bot-token" {
		t.Fatalf("unexpected bot env: %+v", botEnv)
	}

	adapterEnv := applied.Services[2].Env
	if adapterEnv["ADAPTER_INSTANCE_NAME"] != "dahua-adapter" || adapterEnv["DS_BEARER_TOKEN"] != "device-token" {
		t.Fatalf("unexpected adapter env: %+v", adapterEnv)
	}
	if adapterEnv["LOG_DIR"] != filepath.Join(layout.LogsDir, "dahua-terminal-adapter") {
		t.Fatalf("expected adapter log dir, got %+v", adapterEnv)
	}
	if adapterEnv["BACKUP_DIR"] != filepath.Join(layout.BackupsDir, "dahua-terminal-adapter") {
		t.Fatalf("expected adapter backup dir, got %+v", adapterEnv)
	}
}

func TestServiceConfiguredRequiresExpectedEnv(t *testing.T) {
	if ServiceConfigured(ServiceSpec{Name: "bot", Kind: "process-group", Commands: []CommandSpec{{Name: "bot"}}, Env: map[string]string{"BOT_INTERNAL_TOKEN": "x"}}) {
		t.Fatalf("expected bot service to be unconfigured without telegram token")
	}
	if !ServiceConfigured(ServiceSpec{Name: "api", Kind: "process-group", Commands: []CommandSpec{{Name: "api"}}, Env: map[string]string{
		"CORE_TOKEN":       "x",
		"CORE_HMAC_SECRET": "y",
		"ADMIN_JWT_SECRET": "01234567890123456789012345678901",
		"DB_FILE":          "db.sqlite",
	}}) {
		t.Fatalf("expected api service to be configured")
	}
}
