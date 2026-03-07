package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"sg-supervisor/internal/config"
)

func TestStatusesWithHealthReadyProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := NewManager(config.ServiceCatalog{
		Services: []config.ServiceSpec{{
			Name: "api",
			Kind: "process-group",
			HealthChecks: []config.HealthCheckSpec{
				{Name: "api-health", URL: server.URL, TimeoutMS: 1000},
			},
		}},
	})
	manager.statuses["api"] = ServiceStatus{
		Name:       "api",
		Kind:       "process-group",
		State:      "running",
		Components: []ComponentStatus{{Name: "api", State: "running"}},
	}

	statuses := manager.StatusesWithHealth(context.Background())
	if len(statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(statuses))
	}
	if statuses[0].Readiness != "ready" {
		t.Fatalf("expected ready status, got %+v", statuses[0])
	}
}

func TestStatusesWithHealthUnknownWhenNoChecks(t *testing.T) {
	manager := NewManager(config.ServiceCatalog{
		Services: []config.ServiceSpec{{Name: "worker", Kind: "process-group"}},
	})
	manager.statuses["worker"] = ServiceStatus{
		Name:       "worker",
		Kind:       "process-group",
		State:      "running",
		Components: []ComponentStatus{{Name: "worker", State: "running"}},
	}

	statuses := manager.StatusesWithHealth(context.Background())
	if statuses[0].Readiness != "unknown" {
		t.Fatalf("expected unknown readiness, got %+v", statuses[0])
	}
}
