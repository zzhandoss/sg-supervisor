package distribution

import (
	"io/fs"
	"os"
	"path/filepath"
)

func copyTree(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		currentTarget := filepath.Join(targetDir, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(currentTarget, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(currentTarget), 0o755); err != nil {
			return err
		}
		return os.WriteFile(currentTarget, data, info.Mode())
	})
}
