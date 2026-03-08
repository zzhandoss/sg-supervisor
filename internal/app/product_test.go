package app

import (
	"context"
	"os"
	"testing"

	"sg-supervisor/internal/config"
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
	if string(data) != "TELEGRAM_BOT_TOKEN=token-123\n" {
		t.Fatalf("unexpected env file: %q", string(data))
	}
}

func TestUpdateSetupFieldRequiresTelegramBotTokenForCompletedStatus(t *testing.T) {
	supervisor := newTestApp(t)

	if _, err := supervisor.UpdateSetupField(context.Background(), "telegram-bot", "completed", ""); err == nil {
		t.Fatalf("expected missing telegram token error")
	}
}
