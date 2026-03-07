package release

type Report struct {
	Version       string   `json:"version"`
	Platform      string   `json:"platform"`
	ReleaseDir    string   `json:"releaseDir"`
	ArtifactPath  string   `json:"artifactPath"`
	MetadataPath  string   `json:"metadataPath"`
	ChecksumsPath string   `json:"checksumsPath"`
	Warnings      []string `json:"warnings,omitempty"`
}

type SetReport struct {
	Version      string   `json:"version"`
	Platforms    []string `json:"platforms"`
	ReleaseDir   string   `json:"releaseDir"`
	MetadataPath string   `json:"metadataPath"`
	Reports      []Report `json:"reports"`
	Warnings     []string `json:"warnings,omitempty"`
}
