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
		"node.exe":     {},
		"LICENSE":      {},
		"corepack":     {},
		"corepack.cmd": {},
		"node_modules": {},
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, ok := keep[entry.Name()]; ok {
			if entry.Name() == "node_modules" {
				if err := trimNodeModules(filepath.Join(root, "node_modules")); err != nil {
					return err
				}
			}
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func trimNodeModules(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == "corepack" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
