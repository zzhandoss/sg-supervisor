package runtime

type ComponentStatus struct {
	Name       string `json:"name"`
	Executable string `json:"executable,omitempty"`
	State      string `json:"state"`
	PID        int    `json:"pid,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	StoppedAt  string `json:"stoppedAt,omitempty"`
	LastError  string `json:"lastError,omitempty"`
}

type ServiceStatus struct {
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Configured      bool              `json:"configured"`
	RequiresLicense bool              `json:"requiresLicense"`
	State           string            `json:"state"`
	Readiness       string            `json:"readiness"`
	Reachability    string            `json:"reachability"`
	PrimaryURL      string            `json:"primaryUrl,omitempty"`
	StaticDir       string            `json:"staticDir,omitempty"`
	LastError       string            `json:"lastError,omitempty"`
	HealthChecks    []HealthStatus    `json:"healthChecks,omitempty"`
	AccessChecks    []HealthStatus    `json:"accessChecks,omitempty"`
	Components      []ComponentStatus `json:"components,omitempty"`
}
