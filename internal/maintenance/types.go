package maintenance

type Issue struct {
	Step     string `json:"step"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type InstallReport struct {
	PackageID       string   `json:"packageId"`
	ActivePackageID string   `json:"activePackageId"`
	ServiceName     string   `json:"serviceName"`
	WrittenFiles    []string `json:"writtenFiles"`
	InstallHints    []string `json:"installHints"`
}

type RepairReport struct {
	EnsuredPaths        []string `json:"ensuredPaths"`
	ServiceArtifacts    []string `json:"serviceArtifacts"`
	ActivePackageID     string   `json:"activePackageId,omitempty"`
	NeedsPackageInstall bool     `json:"needsPackageInstall"`
}

type UninstallReport struct {
	Mode            string   `json:"mode"`
	Completed       bool     `json:"completed"`
	RemovedPaths    []string `json:"removedPaths"`
	KeptPaths       []string `json:"keptPaths"`
	StoppedServices []string `json:"stoppedServices"`
	UninstallHints  []string `json:"uninstallHints"`
	Issues          []Issue  `json:"issues,omitempty"`
}
