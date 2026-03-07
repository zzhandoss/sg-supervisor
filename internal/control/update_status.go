package control

type UpdateOperationStatus struct {
	Action          string `json:"action"`
	PackageID       string `json:"packageId,omitempty"`
	Outcome         string `json:"outcome"`
	Message         string `json:"message,omitempty"`
	RollbackOutcome string `json:"rollbackOutcome,omitempty"`
	ActivePackageID string `json:"activePackageId,omitempty"`
	RecordedAt      string `json:"recordedAt,omitempty"`
}
