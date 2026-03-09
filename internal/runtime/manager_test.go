package runtime

import (
	"context"
	"os"
	"testing"
	"time"

	"sg-supervisor/internal/config"
)

func TestManagerStartAndStop(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		time.Sleep(10 * time.Second)
		return
	}

	manager := NewManager(config.ServiceCatalog{
		Services: []config.ServiceSpec{
			{
				Name:            "api",
				Kind:            "process-group",
				Env: map[string]string{
					"GO_WANT_HELPER_PROCESS": "1",
					"CORE_TOKEN":             "core-token",
					"CORE_HMAC_SECRET":       "core-hmac",
					"ADMIN_JWT_SECRET":       "01234567890123456789012345678901",
					"DB_FILE":                "db.sqlite",
				},
				RequiresLicense: true,
				Commands: []config.CommandSpec{
					{
						Name:       "api",
						Executable: os.Args[0],
						Args:       []string{"-test.run=TestManagerStartAndStop"},
					},
				},
			},
		},
	})

	if err := manager.Start(context.Background(), "api", true); err != nil {
		t.Fatalf("start: %v", err)
	}
	statuses := manager.Statuses()
	if len(statuses) != 1 || statuses[0].State != "running" {
		t.Fatalf("expected running service, got %+v", statuses)
	}

	if err := manager.Stop("api"); err != nil {
		t.Fatalf("stop: %v", err)
	}

	waitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := manager.WaitForStopped(waitCtx, []string{"api"}); err != nil {
		t.Fatalf("wait for stopped: %v", err)
	}

	statuses = manager.Statuses()
	if statuses[0].State != "error" && statuses[0].State != "stopped" {
		t.Fatalf("expected stopped or error state after stop, got %+v", statuses[0])
	}
}

func TestManagerBlocksLicensedServiceWithoutLicense(t *testing.T) {
	manager := NewManager(config.ServiceCatalog{
		Services: []config.ServiceSpec{{
			Name:            "api",
			Kind:            "process-group",
			RequiresLicense: true,
			Env: map[string]string{
				"CORE_TOKEN":       "core-token",
				"CORE_HMAC_SECRET": "core-hmac",
				"ADMIN_JWT_SECRET": "01234567890123456789012345678901",
				"DB_FILE":          "db.sqlite",
			},
			Commands:        []config.CommandSpec{{Name: "api", Executable: os.Args[0]}},
		}},
	})

	if err := manager.Start(context.Background(), "api", false); err == nil {
		t.Fatalf("expected license gating error")
	}
}
