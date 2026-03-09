package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestProductStorePersistsTelegramBotTokenAndEnvFile(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewProductStore(layout)
	store.hosts = func() []string { return []string{"localhost", "127.0.0.1", "10.20.30.40"} }
	if err := store.SetTelegramBotToken("token-123"); err != nil {
		t.Fatalf("set telegram token: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("load product config: %v", err)
	}
	if cfg.TelegramBotToken != "token-123" {
		t.Fatalf("unexpected product config: %+v", cfg)
	}

	data, err := os.ReadFile(ProductEnvFile(layout))
	if err != nil {
		t.Fatalf("read product env file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "TELEGRAM_BOT_TOKEN=token-123\n") {
		t.Fatalf("unexpected env file: %q", content)
	}
	if !strings.Contains(content, "VITE_API_BASE_URL=http://10.20.30.40:3000\n") {
		t.Fatalf("missing vite api base url: %q", content)
	}
	if !strings.Contains(content, "API_CORS_ALLOWED_ORIGINS=http://10.20.30.40:3000") {
		t.Fatalf("missing api cors origins: %q", content)
	}
}

func TestApplyRuntimeConfigInjectsTelegramBotToken(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{Name: "bot", Kind: "process-group", Env: map[string]string{"TELEGRAM_BOT_TOKEN": "old"}},
			{Name: "api", Kind: "process-group", Env: map[string]string{"NODE_ENV": "production"}},
		},
	}
	internal := InternalRuntimeConfig{
		CoreToken:                "core-token",
		CoreHMACSecret:           "core-hmac",
		AdminJWTSecret:           "01234567890123456789012345678901",
		DeviceServiceToken:       "device-token",
		DeviceServiceInternalKey: "device-internal",
		BotInternalToken:         "bot-internal",
	}

	applied := ApplyRuntimeConfig(layout, catalog, ProductConfig{TelegramBotToken: "token-123"}, internal)
	if got := applied.Services[0].Env["TELEGRAM_BOT_TOKEN"]; got != "token-123" {
		t.Fatalf("expected telegram token override, got %q", got)
	}
	if got := applied.Services[0].Env["SG_PRODUCT_ENV_FILE"]; got != filepath.Join(layout.RuntimeDir, "config", "product.env") {
		t.Fatalf("unexpected product env path: %q", got)
	}
	if _, ok := applied.Services[1].Env["TELEGRAM_BOT_TOKEN"]; ok {
		t.Fatalf("did not expect telegram token on api service")
	}
	if got := applied.Services[1].Env["API_CORS_ALLOWED_ORIGINS"]; got == "" {
		t.Fatalf("expected api cors origins to be injected")
	}
}

func TestApplyRuntimeConfigRemovesTelegramBotTokenWhenUnset(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{Name: "bot", Kind: "process-group", Env: map[string]string{"TELEGRAM_BOT_TOKEN": "old"}},
		},
	}
	internal := InternalRuntimeConfig{
		CoreToken:                "core-token",
		CoreHMACSecret:           "core-hmac",
		AdminJWTSecret:           "01234567890123456789012345678901",
		DeviceServiceToken:       "device-token",
		DeviceServiceInternalKey: "device-internal",
		BotInternalToken:         "bot-internal",
	}

	applied := ApplyRuntimeConfig(layout, catalog, ProductConfig{}, internal)
	if _, ok := applied.Services[0].Env["TELEGRAM_BOT_TOKEN"]; ok {
		t.Fatalf("expected telegram token to be removed when unset")
	}
}

func TestDeriveProductConfigStatusUsesPreferredHostAndMachineOrigins(t *testing.T) {
	status := deriveProductConfigStatus(ProductConfig{
		PreferredHost:    "school-gate.local",
		TelegramBotToken: "token-123",
	}, []string{"localhost", "127.0.0.1", "10.20.30.40"})

	if status.ResolvedHost != "school-gate.local" {
		t.Fatalf("unexpected resolved host: %+v", status)
	}
	if !status.TelegramBotConfigured {
		t.Fatalf("expected telegram bot to be configured")
	}
	if status.ViteAPIBaseURL != "http://school-gate.local:3000" {
		t.Fatalf("unexpected api base url: %+v", status)
	}
	if status.AdminUIURL != "http://school-gate.local:5000" {
		t.Fatalf("unexpected admin ui url: %+v", status)
	}
	if !slices.Contains(status.APICorsAllowedOrigins, "http://10.20.30.40:5000") {
		t.Fatalf("expected machine ip in cors origins: %+v", status.APICorsAllowedOrigins)
	}
}
