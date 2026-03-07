package runtime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"sg-supervisor/internal/config"
)

type HealthStatus struct {
	Name          string `json:"name"`
	URL           string `json:"url,omitempty"`
	Readiness     string `json:"readiness"`
	Message       string `json:"message,omitempty"`
	LastCheckedAt string `json:"lastCheckedAt,omitempty"`
}

func (m *Manager) StatusesWithHealth(ctx context.Context) []ServiceStatus {
	m.mu.Lock()
	statuses := make([]ServiceStatus, 0, len(m.order))
	for _, name := range m.orderedNames() {
		statuses = append(statuses, m.statuses[name])
	}
	specs := make(map[string]config.ServiceSpec, len(m.specs))
	for name, spec := range m.specs {
		specs[name] = spec
	}
	m.mu.Unlock()

	enriched := make([]ServiceStatus, 0, len(statuses))
	for _, status := range statuses {
		enriched = append(enriched, evaluateReadiness(ctx, status, specs[status.Name]))
	}
	return enriched
}

func evaluateReadiness(ctx context.Context, status ServiceStatus, spec config.ServiceSpec) ServiceStatus {
	status.HealthChecks = nil
	status.Readiness = "unknown"

	if spec.Kind == "static-assets" {
		if spec.StaticDir == "" {
			status.Readiness = "not_ready"
			status.HealthChecks = []HealthStatus{{Name: "static-assets", Readiness: "not_ready", Message: "staticDir is not configured"}}
			return status
		}
		if _, err := os.Stat(spec.StaticDir); err != nil {
			status.Readiness = "not_ready"
			status.HealthChecks = []HealthStatus{{Name: "static-assets", Readiness: "not_ready", Message: err.Error()}}
			return status
		}
		status.Readiness = "ready"
		status.HealthChecks = []HealthStatus{{Name: "static-assets", Readiness: "ready", Message: "static assets are present"}}
		return status
	}

	if status.State != "running" {
		status.Readiness = "not_ready"
		status.HealthChecks = []HealthStatus{{Name: "process-group", Readiness: "not_ready", Message: "service is not running"}}
		return status
	}

	if len(spec.HealthChecks) == 0 {
		status.Readiness = "unknown"
		status.HealthChecks = []HealthStatus{{Name: "health-checks", Readiness: "unknown", Message: "no health checks configured"}}
		return status
	}

	ready := true
	now := time.Now().UTC().Format(time.RFC3339)
	for _, probe := range spec.HealthChecks {
		check := runHTTPHealthCheck(ctx, probe, now)
		if check.Readiness != "ready" {
			ready = false
		}
		status.HealthChecks = append(status.HealthChecks, check)
	}
	if ready {
		status.Readiness = "ready"
	} else {
		status.Readiness = "not_ready"
	}
	return status
}

func runHTTPHealthCheck(ctx context.Context, probe config.HealthCheckSpec, checkedAt string) HealthStatus {
	timeout := 3 * time.Second
	if probe.TimeoutMS > 0 {
		timeout = time.Duration(probe.TimeoutMS) * time.Millisecond
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, probe.URL, nil)
	if err != nil {
		return HealthStatus{Name: probe.Name, URL: probe.URL, Readiness: "not_ready", Message: err.Error(), LastCheckedAt: checkedAt}
	}
	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return HealthStatus{Name: probe.Name, URL: probe.URL, Readiness: "not_ready", Message: err.Error(), LastCheckedAt: checkedAt}
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return HealthStatus{Name: probe.Name, URL: probe.URL, Readiness: "ready", Message: fmt.Sprintf("status %d", response.StatusCode), LastCheckedAt: checkedAt}
	}
	return HealthStatus{Name: probe.Name, URL: probe.URL, Readiness: "not_ready", Message: fmt.Sprintf("status %d", response.StatusCode), LastCheckedAt: checkedAt}
}
