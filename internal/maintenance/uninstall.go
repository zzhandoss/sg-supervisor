package maintenance

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"sg-supervisor/internal/config"
)

const (
	ModeKeepState = "keep-state"
	ModeFullWipe  = "full-wipe"
)

func ExecuteUninstall(layout config.Layout, mode string) (UninstallReport, error) {
	if mode == "" {
		mode = ModeKeepState
	}
	if mode != ModeKeepState && mode != ModeFullWipe {
		return UninstallReport{}, errors.New("unsupported uninstall mode")
	}

	var removePaths []string
	var keepPaths []string

	switch mode {
	case ModeKeepState:
		removePaths = []string{
			layout.InstallDir,
			layout.RuntimeDir,
			layout.UpdatesDir,
		}
		keepPaths = []string{
			layout.ConfigDir,
			layout.DataDir,
			layout.LogsDir,
			layout.LicensesDir,
			layout.BackupsDir,
		}
	case ModeFullWipe:
		removePaths = []string{
			layout.InstallDir,
			layout.RuntimeDir,
			layout.UpdatesDir,
			layout.ConfigDir,
			layout.DataDir,
			layout.LogsDir,
			layout.LicensesDir,
			layout.BackupsDir,
		}
	}

	removed := make([]string, 0, len(removePaths))
	report := UninstallReport{
		Mode:      mode,
		KeptPaths: uniquePaths(keepPaths),
	}
	for _, path := range uniquePaths(removePaths) {
		if err := os.RemoveAll(path); err != nil {
			report.RemovedPaths = removed
			return report, err
		}
		removed = append(removed, path)
	}

	report.Completed = true
	report.RemovedPaths = removed
	return report, nil
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	sort.Strings(result)
	return result
}
