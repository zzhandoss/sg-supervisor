package runtime

import (
	"context"
	"fmt"

	"sg-supervisor/internal/config"
)

func ApplyReachability(ctx context.Context, statuses []ServiceStatus, host string) []ServiceStatus {
	enriched := make([]ServiceStatus, 0, len(statuses))
	for _, status := range statuses {
		enriched = append(enriched, evaluateReachability(ctx, status, host))
	}
	return enriched
}

func evaluateReachability(ctx context.Context, status ServiceStatus, host string) ServiceStatus {
	status.AccessChecks = nil
	status.Reachability = "unknown"

	probes := externalAccessChecks(status.Name, host)
	if len(probes) == 0 {
		return status
	}

	status.PrimaryURL = primaryURLForChecks(probes)
	if !status.Configured {
		status.Reachability = "not_ready"
		status.AccessChecks = []HealthStatus{{
			Name:      "reachability",
			URL:       status.PrimaryURL,
			Readiness: "not_ready",
			Message:   "service is not configured",
		}}
		return status
	}
	if status.State != "running" {
		status.Reachability = "not_ready"
		status.AccessChecks = []HealthStatus{{
			Name:      "reachability",
			URL:       status.PrimaryURL,
			Readiness: "not_ready",
			Message:   "service process is not running",
		}}
		return status
	}

	ready := true
	for _, probe := range probes {
		check := runHTTPHealthCheck(ctx, probe, "")
		if check.Readiness != "ready" {
			ready = false
		}
		status.AccessChecks = append(status.AccessChecks, check)
	}
	if ready {
		status.Reachability = "ready"
		return status
	}
	status.Reachability = "not_ready"
	return status
}

func externalAccessChecks(serviceName, host string) []config.HealthCheckSpec {
	if host == "" {
		host = "127.0.0.1"
	}
	switch serviceName {
	case "api":
		return []config.HealthCheckSpec{{Name: "api-access", URL: fmt.Sprintf("http://%s:3000/health", host), TimeoutMS: 3000}}
	case "device-service":
		return []config.HealthCheckSpec{{Name: "device-service-access", URL: fmt.Sprintf("http://%s:4010/health", host), TimeoutMS: 3000}}
	case "admin-ui":
		return []config.HealthCheckSpec{{Name: "admin-ui-access", URL: fmt.Sprintf("http://%s:5000/healthz", host), TimeoutMS: 3000}}
	case "dahua-terminal-adapter":
		return []config.HealthCheckSpec{{Name: "adapter-access", URL: fmt.Sprintf("http://%s:8091/monitor/snapshot", host), TimeoutMS: 3000}}
	default:
		return nil
	}
}

func primaryURLForChecks(checks []config.HealthCheckSpec) string {
	if len(checks) == 0 {
		return ""
	}
	return checks[0].URL
}
