package runtime

import (
	"context"
	"os"
	"testing"
	"time"

	"sg-supervisor/internal/config"
)

func TestManagerReconfigureUpdatesServiceEnv(t *testing.T) {
	if os.Getenv("GO_WANT_RECONFIGURE_HELPER") == "1" {
		if os.Getenv("TELEGRAM_BOT_TOKEN") != "token-123" {
			os.Exit(2)
		}
		time.Sleep(10 * time.Second)
		return
	}

	manager := NewManager(config.ServiceCatalog{
		Services: []config.ServiceSpec{{
			Name: "bot",
			Kind: "process-group",
			Env: map[string]string{
				"GO_WANT_RECONFIGURE_HELPER": "1",
				"BOT_INTERNAL_TOKEN":         "bot-internal",
			},
			Commands: []config.CommandSpec{{
				Name:       "bot",
				Executable: os.Args[0],
				Args:       []string{"-test.run=TestManagerReconfigureUpdatesServiceEnv"},
			}},
		}},
	})

	manager.Reconfigure(config.ServiceCatalog{
		Services: []config.ServiceSpec{{
			Name: "bot",
			Kind: "process-group",
			Env: map[string]string{
				"GO_WANT_RECONFIGURE_HELPER": "1",
				"BOT_INTERNAL_TOKEN":         "bot-internal",
				"TELEGRAM_BOT_TOKEN":         "token-123",
			},
			Commands: []config.CommandSpec{{
				Name:       "bot",
				Executable: os.Args[0],
				Args:       []string{"-test.run=TestManagerReconfigureUpdatesServiceEnv"},
			}},
		}},
	})

	if err := manager.Start(context.Background(), "bot", true); err != nil {
		t.Fatalf("start after reconfigure: %v", err)
	}
	time.Sleep(150 * time.Millisecond)
	statuses := manager.Statuses()
	if len(statuses) != 1 || statuses[0].State != "running" {
		t.Fatalf("expected running bot after reconfigure, got %+v", statuses)
	}
	_ = manager.Stop("bot")
}
