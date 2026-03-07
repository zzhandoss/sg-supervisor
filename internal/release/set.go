package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

func BuildSet(root, version string, reports []Report) (SetReport, error) {
	version = normalizeVersion(version)
	releaseDir := filepath.Join(root, "releases", "v"+version)
	platforms := make([]string, 0, len(reports))
	warnings := make([]string, 0)

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Platform < reports[j].Platform
	})
	for _, report := range reports {
		platforms = append(platforms, report.Platform)
		warnings = append(warnings, report.Warnings...)
	}

	metadataPath, err := writeSetMetadata(releaseDir, version, reports, warnings)
	if err != nil {
		return SetReport{}, err
	}
	return SetReport{
		Version:      version,
		Platforms:    platforms,
		ReleaseDir:   releaseDir,
		MetadataPath: metadataPath,
		Reports:      reports,
		Warnings:     warnings,
	}, nil
}

func writeSetMetadata(releaseDir, version string, reports []Report, warnings []string) (string, error) {
	body := map[string]any{
		"version":   version,
		"reports":   reports,
		"warnings":  warnings,
		"platforms": collectPlatforms(reports),
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(releaseDir, "release-set.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func collectPlatforms(reports []Report) []string {
	platforms := make([]string, 0, len(reports))
	for _, report := range reports {
		platforms = append(platforms, report.Platform)
	}
	sort.Strings(platforms)
	return platforms
}
