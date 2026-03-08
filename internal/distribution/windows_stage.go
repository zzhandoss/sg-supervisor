package distribution

import (
	"io/fs"
	"os"
	"path/filepath"
)

func prepareWindowsBuildStage(sourceDir string) (string, func(), error) {
	baseRoot := filepath.Join(os.TempDir(), "s")
	if err := os.MkdirAll(baseRoot, 0o755); err != nil {
		return "", nil, err
	}
	root, err := os.MkdirTemp(baseRoot, "w-")
	if err != nil {
		return "", nil, err
	}
	if err := copyWindowsBuildStage(sourceDir, root); err != nil {
		_ = os.RemoveAll(root)
		return "", nil, err
	}
	return root, func() {
		_ = os.RemoveAll(root)
	}, nil
}

func copyWindowsBuildStage(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(targetDir, 0o755)
		}
		targetPath := filepath.Join(targetDir, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}
