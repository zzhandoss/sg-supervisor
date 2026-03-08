package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProductStorePersistsTelegramBotTokenAndEnvFile(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewProductStore(layout)
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
	if string(data) != "TELEGRAM_BOT_TOKEN=token-123\n" {
		t.Fatalf("unexpected env file: %q", string(data))
	}
}

func TestApplyProductConfigInjectsTelegramBotToken(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{Name: "bot", Kind: "process-group", Env: map[string]string{"TELEGRAM_BOT_TOKEN": "old"}},
			{Name: "api", Kind: "process-group", Env: map[string]string{"NODE_ENV": "production"}},
		},
	}

	applied := ApplyProductConfig(layout, catalog, ProductConfig{TelegramBotToken: "token-123"})
	if got := applied.Services[0].Env["TELEGRAM_BOT_TOKEN"]; got != "token-123" {
		t.Fatalf("expected telegram token override, got %q", got)
	}
	if got := applied.Services[0].Env["SG_PRODUCT_ENV_FILE"]; got != filepath.Join(layout.RuntimeDir, "config", "product.env") {
		t.Fatalf("unexpected product env path: %q", got)
	}
	if _, ok := applied.Services[1].Env["TELEGRAM_BOT_TOKEN"]; ok {
		t.Fatalf("did not expect telegram token on api service")
	}
}

func TestApplyProductConfigRemovesTelegramBotTokenWhenUnset(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{Name: "bot", Kind: "process-group", Env: map[string]string{"TELEGRAM_BOT_TOKEN": "old"}},
		},
	}

	applied := ApplyProductConfig(layout, catalog, ProductConfig{})
	if _, ok := applied.Services[0].Env["TELEGRAM_BOT_TOKEN"]; ok {
		t.Fatalf("expected telegram token to be removed when unset")
	}
}
