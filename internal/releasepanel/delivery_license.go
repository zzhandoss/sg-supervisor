package releasepanel

import (
	"os"
	"path/filepath"
)

func (s *Service) latestIssuedFreeLicensePath() (string, error) {
	records, err := s.store.ListIssuedLicenses()
	if err != nil {
		return "", err
	}
	for _, record := range records {
		if record.Mode == "free" && record.Path != "" {
			return record.Path, nil
		}
	}
	return "", nil
}

func copyBundledFreeLicense(root, sourcePath string) (string, error) {
	if sourcePath == "" {
		return "", nil
	}
	targetDir := filepath.Join(root, "licenses")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	targetPath := filepath.Join(targetDir, "free-license.json")
	if err := copyArtifact(sourcePath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}
