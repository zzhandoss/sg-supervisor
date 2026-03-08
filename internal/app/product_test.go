package app

import (
	"context"
	"os"
	"strings"
	"testing"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/control"
)

func TestUpdateSetupFieldPersistsTelegramBotConfig(t *testing.T) {
	supervisor := newTestApp(t)

	status, err := supervisor.UpdateSetupField(context.Background(), "telegram-bot", "completed", "token-123")
	if err != nil {
		t.Fatalf("update setup field: %v", err)
	}
	if status.Complete {
		t.Fatalf("expected setup to remain incomplete without license")
	}

	productCfg, err := supervisor.product.Load()
	if err != nil {
		t.Fatalf("load product config: %v", err)
	}
	if productCfg.TelegramBotToken != "token-123" {
		t.Fatalf("unexpected product config: %+v", productCfg)
	}

	data, err := os.ReadFile(config.ProductEnvFile(supervisor.layout))
	if err != nil {
		t.Fatalf("read product env file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "TELEGRAM_BOT_TOKEN=token-123\n") {
		t.Fatalf("unexpected env file: %q", content)
	}
	if !strings.Contains(content, "VITE_API_BASE_URL=") {
		t.Fatalf("expected api base url in env file: %q", content)
	}
}

func TestUpdateSetupFieldRequiresTelegramBotTokenForCompletedStatus(t *testing.T) {
	supervisor := newTestApp(t)

	if _, err := supervisor.UpdateSetupField(context.Background(), "telegram-bot", "completed", ""); err == nil {
		t.Fatalf("expected missing telegram token error")
	}
}

func TestUpdateSetupFieldUsesExistingTelegramBotTokenWhenPresent(t *testing.T) {
	supervisor := newTestApp(t)
	supervisor.product.Save(config.ProductConfig{TelegramBotToken: "token-123"})

	status, err := supervisor.UpdateSetupField(context.Background(), "telegram-bot", "completed", "")
	if err != nil {
		t.Fatalf("update setup field: %v", err)
	}
	if len(status.Optional) == 0 || status.Optional[0].Status != "completed" {
		t.Fatalf("expected telegram-bot to be marked completed: %+v", status)
	}
}

func TestUpdateProductConfigPersistsPreferredHostAndCompletesTelegramSetup(t *testing.T) {
	supervisor := newTestApp(t)

	preferredHost := "10.20.30.40"
	token := "token-123"
	status, err := supervisor.UpdateProductConfig(context.Background(), control.ProductConfigUpdate{
		PreferredHost:    &preferredHost,
		TelegramBotToken: &token,
	})
	if err != nil {
		t.Fatalf("update product config: %v", err)
	}
	if status.ResolvedHost != "10.20.30.40" || !status.TelegramBotConfigured {
		t.Fatalf("unexpected product config status: %+v", status)
	}
}
