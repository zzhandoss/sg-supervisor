package releasepanel

import (
	"io/fs"
	"os"
	"path/filepath"
)

func copyMaterializedDir(sourceDir, targetDir string) error {
	return copyMaterializedPath(sourceDir, targetDir, map[string]bool{})
}

func copyMaterializedPath(sourcePath, targetPath string, stack map[string]bool) error {
	resolvedPath, info, err := resolveMaterializedPath(sourcePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		key := filepath.Clean(resolvedPath)
		if stack[key] {
			return nil
		}
		stack[key] = true
		defer delete(stack, key)
		if err := os.MkdirAll(targetPath, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(resolvedPath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyMaterializedPath(filepath.Join(resolvedPath, entry.Name()), filepath.Join(targetPath, entry.Name()), stack); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, data, info.Mode())
}

func resolveMaterializedPath(path string) (string, fs.FileInfo, error) {
	resolvedPath := path
	if realPath, err := filepath.EvalSymlinks(path); err == nil {
		resolvedPath = realPath
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", nil, err
	}
	return resolvedPath, info, nil
}
