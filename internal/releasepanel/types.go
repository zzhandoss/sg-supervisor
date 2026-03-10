package releasepanel

const (
	RepoSchoolGate = "zzhandoss/school-gate"
	RepoAdapter    = "zzhandoss/dahua-terminal-adapter"
	RepoNode       = "nodejs/node"
	RepoWinSW      = "winsw/winsw"
)

const winSWVersion = "v2.12.0"

const (
	JobTypeLocalRelease = "local-release"
	JobStatusQueued     = "queued"
	JobStatusRunning    = "running"
	JobStatusSucceeded  = "succeeded"
	JobStatusFailed     = "failed"
)

type Recipe struct {
	InstallerVersion  string `json:"installerVersion"`
	SchoolGateVersion string `json:"schoolGateVersion"`
	AdapterVersion    string `json:"adapterVersion"`
	NodeVersion       string `json:"nodeVersion"`
}

type SigningKeys struct {
	LicensePrivateKeyBase64 string `json:"licensePrivateKeyBase64,omitempty"`
	LicensePublicKeyBase64  string `json:"licensePublicKeyBase64,omitempty"`
	PackagePrivateKeyBase64 string `json:"packagePrivateKeyBase64,omitempty"`
	PackagePublicKeyBase64  string `json:"packagePublicKeyBase64,omitempty"`
}

type State struct {
	ListenAddress string      `json:"listenAddress"`
	RepoRoot      string      `json:"repoRoot"`
	Recipe        Recipe      `json:"recipe"`
	Keys          SigningKeys `json:"keys"`
}

type KeysStatus struct {
	LicensePublicKeyBase64 string `json:"licensePublicKeyBase64,omitempty"`
	PackagePublicKeyBase64 string `json:"packagePublicKeyBase64,omitempty"`
	LicenseConfigured      bool   `json:"licenseConfigured"`
	PackageConfigured      bool   `json:"packageConfigured"`
}

type Status struct {
	Root           string                `json:"root"`
	ListenAddress  string                `json:"listenAddress"`
	RepoRoot       string                `json:"repoRoot"`
	HostPlatform   string                `json:"hostPlatform"`
	Recipe         Recipe                `json:"recipe"`
	Keys           KeysStatus            `json:"keys"`
	Jobs           []Job                 `json:"jobs"`
	IssuedLicenses []IssuedLicenseRecord `json:"issuedLicenses"`
	ReleaseDir     string                `json:"releaseDir"`
}

type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	CreatedAt  string                 `json:"createdAt"`
	StartedAt  string                 `json:"startedAt,omitempty"`
	FinishedAt string                 `json:"finishedAt,omitempty"`
	Recipe     Recipe                 `json:"recipe"`
	Logs       []string               `json:"logs,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Report     *LocalReleaseSetReport `json:"report,omitempty"`
}

type LocalReleaseReport struct {
	Version       string   `json:"version"`
	Platform      string   `json:"platform"`
	ReleaseDir    string   `json:"releaseDir"`
	ArtifactPath  string   `json:"artifactPath"`
	MetadataPath  string   `json:"metadataPath"`
	ChecksumsPath string   `json:"checksumsPath"`
	Warnings      []string `json:"warnings,omitempty"`
}

type LocalReleaseSetReport struct {
	Version      string               `json:"version"`
	Platforms    []string             `json:"platforms"`
	ReleaseDir   string               `json:"releaseDir"`
	MetadataPath string               `json:"metadataPath"`
	Reports      []LocalReleaseReport `json:"reports"`
	Warnings     []string             `json:"warnings,omitempty"`
}

type ReleaseVersion struct {
	Tag         string `json:"tag"`
	Name        string `json:"name,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
}

type IssuedLicenseRecord struct {
	LicenseID   string   `json:"licenseId"`
	Path        string   `json:"path"`
	Customer    string   `json:"customer"`
	Mode        string   `json:"mode"`
	Edition     string   `json:"edition"`
	Features    []string `json:"features,omitempty"`
	ExpiresAt   string   `json:"expiresAt,omitempty"`
	IssuedAt    string   `json:"issuedAt"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

type LicenseIssueRequest struct {
	ActivationRequestPath string   `json:"activationRequestPath,omitempty"`
	Customer              string   `json:"customer,omitempty"`
	Mode                  string   `json:"mode"`
	Edition               string   `json:"edition"`
	Features              []string `json:"features,omitempty"`
	ExpiresAt             string   `json:"expiresAt,omitempty"`
	Perpetual             bool     `json:"perpetual"`
	Fingerprint           string   `json:"fingerprint,omitempty"`
}
