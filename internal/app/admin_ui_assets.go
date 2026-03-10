package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func resolveAdminUIBuildDir(appDir string) (string, error) {
	nitroRoot := filepath.Join(appDir, ".output")
	nitroPublic := filepath.Join(nitroRoot, "public")
	nitroServer := filepath.Join(nitroRoot, "server", "index.mjs")
	if info, err := os.Stat(nitroPublic); err == nil && info.IsDir() {
		if serverInfo, err := os.Stat(nitroServer); err == nil && !serverInfo.IsDir() {
			return nitroRoot, nil
		}
	}
	distRoot := filepath.Join(appDir, "dist")
	distIndex := filepath.Join(distRoot, "index.html")
	if info, err := os.Stat(distIndex); err == nil && !info.IsDir() {
		return distRoot, nil
	}
	return "", errors.New("admin-ui build output is missing: checked " + strings.Join([]string{
		filepath.Join(appDir, ".output", "public"),
		filepath.Join(appDir, ".output", "server", "index.mjs"),
		distIndex,
	}, ", "))
}
