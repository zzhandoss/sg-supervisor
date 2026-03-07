package updates

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func (s *Store) applyBundle(ctx context.Context, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if record.StageDir == "" {
		return errors.New("bundle stage directory is missing")
	}

	payloadRoot := filepath.Join(record.StageDir, "payload")
	if _, err := os.Stat(payloadRoot); os.IsNotExist(err) {
		return errors.New("bundle payload directory is missing")
	} else if err != nil {
		return err
	}

	backupRoot := filepath.Join(s.layout.BackupsDir, record.PackageID)
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return err
	}
	plan := rollbackPlan{}

	targets := map[string]string{
		filepath.Join(payloadRoot, "core"):    filepath.Join(s.layout.InstallDir, "core"),
		filepath.Join(payloadRoot, "runtime"): filepath.Join(s.layout.InstallDir, "runtime"),
	}

	for sourcePath, targetPath := range targets {
		target, err := s.replaceIfPresent(sourcePath, targetPath, backupRoot)
		if err != nil {
			return err
		}
		if target.TargetPath != "" {
			plan.Targets = append(plan.Targets, target)
		}
	}

	adaptersRoot := filepath.Join(payloadRoot, "adapters")
	entries, err := os.ReadDir(adaptersRoot)
	if os.IsNotExist(err) {
		return writeRollbackPlan(filepath.Join(backupRoot, "rollback-plan.json"), plan)
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sourcePath := filepath.Join(adaptersRoot, entry.Name())
		targetPath := filepath.Join(s.layout.InstallDir, "adapters", entry.Name())
		target, err := s.replaceIfPresent(sourcePath, targetPath, filepath.Join(backupRoot, "adapters"))
		if err != nil {
			return err
		}
		if target.TargetPath != "" {
			plan.Targets = append(plan.Targets, target)
		}
	}
	return writeRollbackPlan(filepath.Join(backupRoot, "rollback-plan.json"), plan)
}

func (s *Store) replaceIfPresent(sourcePath, targetPath, backupRoot string) (rollbackTarget, error) {
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return rollbackTarget{}, nil
	} else if err != nil {
		return rollbackTarget{}, err
	}

	target := rollbackTarget{TargetPath: targetPath}
	if _, err := os.Stat(targetPath); err == nil {
		backupPath := filepath.Join(backupRoot, filepath.Base(targetPath))
		if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
			return rollbackTarget{}, err
		}
		if err := os.RemoveAll(backupPath); err != nil {
			return rollbackTarget{}, err
		}
		if err := os.Rename(targetPath, backupPath); err != nil {
			return rollbackTarget{}, err
		}
		target.BackupPath = backupPath
	} else if !os.IsNotExist(err) {
		return rollbackTarget{}, err
	} else {
		target.RemoveOnRollback = true
	}

	return target, copyDir(sourcePath, targetPath)
}

func copyDir(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, rel)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		sourceInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, sourceInfo.Mode())
	})
}
