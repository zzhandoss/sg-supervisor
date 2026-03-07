package distribution

type Report struct {
	Platform       string   `json:"platform"`
	StageDir       string   `json:"stageDir"`
	OutputDir      string   `json:"outputDir"`
	ArtifactPath   string   `json:"artifactPath,omitempty"`
	GeneratedFiles []string `json:"generatedFiles"`
	Warnings       []string `json:"warnings,omitempty"`
}
