package bootstrap

type StepStatus struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	Message    string `json:"message,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
}

type Status struct {
	State              string       `json:"state"`
	CurrentStep        string       `json:"currentStep,omitempty"`
	StartedAt          string       `json:"startedAt,omitempty"`
	FinishedAt         string       `json:"finishedAt,omitempty"`
	Error              string       `json:"error,omitempty"`
	SourceArchivePath  string       `json:"sourceArchivePath,omitempty"`
	AdapterArchivePath string       `json:"adapterArchivePath,omitempty"`
	SourceRoot         string       `json:"sourceRoot,omitempty"`
	Logs               []string     `json:"logs,omitempty"`
	Steps              []StepStatus `json:"steps,omitempty"`
}
