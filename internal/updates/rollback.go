package updates

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type rollbackTarget struct {
	TargetPath       string `json:"targetPath"`
	BackupPath       string `json:"backupPath,omitempty"`
	RemoveOnRollback bool   `json:"removeOnRollback,omitempty"`
}

type rollbackPlan struct {
	Targets []rollbackTarget `json:"targets"`
}

func (s *Store) Rollback(ctx context.Context, backupRoot string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	plan, err := loadRollbackPlan(filepath.Join(backupRoot, "rollback-plan.json"))
	if err != nil {
		return err
	}
	for _, target := range plan.Targets {
		if err := os.RemoveAll(target.TargetPath); err != nil {
			return err
		}
		if target.BackupPath == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target.TargetPath), 0o755); err != nil {
			return err
		}
		if err := os.Rename(target.BackupPath, target.TargetPath); err != nil {
			return err
		}
	}
	return nil
}

func writeRollbackPlan(path string, plan rollbackPlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func loadRollbackPlan(path string) (rollbackPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return rollbackPlan{}, err
	}
	var plan rollbackPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return rollbackPlan{}, err
	}
	return plan, nil
}
