package config

import (
	"os"
	"path/filepath"
)

func EnsureRuntimeDirectories(layout Layout) error {
	dirs := []string{
		filepath.Join(layout.DataDir, "school-gate"),
		filepath.Join(layout.DataDir, "dahua-terminal-adapter"),
		filepath.Join(layout.LogsDir, "dahua-terminal-adapter"),
		filepath.Join(layout.RuntimeDir, "config"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
