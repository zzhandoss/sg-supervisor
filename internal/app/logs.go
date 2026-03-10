package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type RecentLogs struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

func (a *App) ReadRecentLogs(ctx context.Context, limit int) (RecentLogs, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return RecentLogs{}, err
	}
	if limit <= 0 {
		limit = 80
	}
	path := filepath.Join(a.layout.LogsDir, "sg-supervisor.log")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RecentLogs{Path: path, Lines: []string{}}, nil
	}
	if err != nil {
		return RecentLogs{}, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}
	return RecentLogs{Path: path, Lines: filtered}, nil
}
