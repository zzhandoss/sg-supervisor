package runtime

import (
	"context"
	"fmt"
	"time"
)

func (m *Manager) RunningServiceNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	names := make([]string, 0, len(m.statuses))
	for _, name := range m.orderedNames() {
		if m.statuses[name].State == "running" {
			names = append(names, name)
		}
	}
	return names
}

func (m *Manager) WaitForStopped(ctx context.Context, names []string) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if m.allStopped(names) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *Manager) StopMany(names []string) error {
	var stopErr error
	for _, name := range names {
		if err := m.Stop(name); err != nil {
			stopErr = joinErrors(stopErr, fmt.Errorf("%s: %w", name, err))
		}
	}
	return stopErr
}

func (m *Manager) allStopped(names []string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, name := range names {
		state := m.statuses[name].State
		if state == "running" || state == "starting" {
			return false
		}
	}
	return true
}

func joinErrors(current error, next error) error {
	if current == nil {
		return next
	}
	return fmt.Errorf("%v; %w", current, next)
}
