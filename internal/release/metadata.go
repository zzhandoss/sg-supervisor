package release

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"sg-supervisor/internal/distribution"
)

func writeChecksums(releaseDir, artifactPath string) (string, error) {
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	content := hex.EncodeToString(sum[:]) + "  " + filepath.Base(artifactPath) + "\n"
	path := filepath.Join(releaseDir, "SHA256SUMS.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func writeMetadata(releaseDir, version string, report distribution.Report, artifactPath, checksumsPath string, warnings []string) (string, error) {
	body := map[string]any{
		"version":       version,
		"platform":      report.Platform,
		"stageDir":      report.StageDir,
		"distribution":  report.OutputDir,
		"artifactPath":  artifactPath,
		"checksumsPath": checksumsPath,
		"warnings":      warnings,
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(releaseDir, "release.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
