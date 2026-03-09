package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type bootstrapAssets struct {
	SourceArchivePath  string
	AdapterArchivePath string
	SourceRoot         string
}

func locateBootstrapAssets(root, workDir string) (bootstrapAssets, error) {
	payloadDir := filepath.Join(root, "payload")
	entries, err := os.ReadDir(payloadDir)
	if err != nil {
		return bootstrapAssets{}, err
	}
	var sourceMatches []string
	var adapterMatches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		path := filepath.Join(payloadDir, entry.Name())
		if strings.Contains(name, "school-gate-") && strings.Contains(name, "-source") && strings.HasSuffix(name, ".zip") {
			sourceMatches = append(sourceMatches, path)
		}
		if strings.Contains(name, "dahua-adapter-") && strings.HasSuffix(name, ".zip") {
			platformToken := "-linux-x64.zip"
			if runtime.GOOS == "windows" {
				platformToken = "-win-x64.zip"
			}
			if strings.HasSuffix(name, platformToken) {
				adapterMatches = append(adapterMatches, path)
			}
		}
	}
	sort.Strings(sourceMatches)
	sort.Strings(adapterMatches)
	if len(sourceMatches) != 1 {
		return bootstrapAssets{}, errors.New("expected exactly one school-gate source archive in payload")
	}
	if len(adapterMatches) != 1 {
		return bootstrapAssets{}, errors.New("expected exactly one adapter archive in payload")
	}
	return bootstrapAssets{
		SourceArchivePath:  sourceMatches[0],
		AdapterArchivePath: adapterMatches[0],
		SourceRoot:         filepath.Join(workDir, "source"),
	}, nil
}

func detectPackageManager(sourceRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(sourceRoot, "package.json"))
	if err != nil {
		return "", err
	}
	var body struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return "", err
	}
	if strings.TrimSpace(body.PackageManager) == "" {
		return "", errors.New("packageManager is not declared in school-gate package.json")
	}
	return body.PackageManager, nil
}

func nodeExecutablePath(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "install", "runtime", "node", "node.exe")
	}
	return filepath.Join(root, "install", "runtime", "node", "bin", "node")
}

func corepackExecutablePath(root string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(root, "install", "runtime", "node", "corepack.cmd")
	}
	return filepath.Join(root, "install", "runtime", "node", "bin", "corepack")
}

func bootstrapCommandEnv(root string) map[string]string {
	nodeDir := filepath.Dir(nodeExecutablePath(root))
	if runtime.GOOS == "windows" {
		nodeDir = filepath.Join(root, "install", "runtime", "node")
	}
	return map[string]string{
		"PATH":      nodeDir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"PNPM_HOME": nodeDir,
	}
}
