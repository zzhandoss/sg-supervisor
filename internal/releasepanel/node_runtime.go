package releasepanel

import (
	"os"
	"path/filepath"
)

func prepareNodeRuntime(root, platform string) error {
	if platform != "windows" {
		return nil
	}
	keep := map[string]struct{}{
		"node.exe": {},
		"LICENSE":  {},
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, ok := keep[entry.Name()]; ok {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
