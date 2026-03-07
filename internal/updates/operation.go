package updates

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type OperationStatus struct {
	Action          string `json:"action"`
	PackageID       string `json:"packageId,omitempty"`
	Outcome         string `json:"outcome"`
	Message         string `json:"message,omitempty"`
	RollbackOutcome string `json:"rollbackOutcome,omitempty"`
	ActivePackageID string `json:"activePackageId,omitempty"`
	RecordedAt      string `json:"recordedAt"`
}

func (s *Store) SaveOperation(status OperationStatus) error {
	if status.RecordedAt == "" {
		status.RecordedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.operationPath(), data, 0o644)
}

func (s *Store) Operation(ctx context.Context) (OperationStatus, error) {
	if err := ctx.Err(); err != nil {
		return OperationStatus{}, err
	}
	data, err := os.ReadFile(s.operationPath())
	if os.IsNotExist(err) {
		return OperationStatus{}, nil
	}
	if err != nil {
		return OperationStatus{}, err
	}
	var status OperationStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return OperationStatus{}, err
	}
	return status, nil
}

func (s *Store) operationPath() string {
	return filepath.Join(s.layout.UpdatesDir, "last-operation.json")
}
