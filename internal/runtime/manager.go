package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"sync"
	"time"

	"sg-supervisor/internal/config"
)

type processEntry struct {
	name    string
	command *exec.Cmd
}

type Manager struct {
	mu        sync.Mutex
	order     []string
	specs     map[string]config.ServiceSpec
	processes map[string][]processEntry
	statuses  map[string]ServiceStatus
}

func NewManager(catalog config.ServiceCatalog) *Manager {
	manager := &Manager{
		specs:     make(map[string]config.ServiceSpec, len(catalog.Services)),
		processes: make(map[string][]processEntry, len(catalog.Services)),
		statuses:  make(map[string]ServiceStatus, len(catalog.Services)),
	}

	for _, service := range catalog.Services {
		manager.order = append(manager.order, service.Name)
		manager.specs[service.Name] = service
		manager.statuses[service.Name] = buildInitialStatus(service)
	}
	return manager
}

func (m *Manager) Statuses() []ServiceStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	statuses := make([]ServiceStatus, 0, len(m.order))
	for _, name := range m.orderedNames() {
		statuses = append(statuses, m.statuses[name])
	}
	return statuses
}

func (m *Manager) Start(ctx context.Context, name string, licenseValid bool) error {
	m.mu.Lock()
	spec, ok := m.specs[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("unknown service %q", name)
	}
	if spec.RequiresLicense && !licenseValid {
		m.mu.Unlock()
		return errors.New("cannot start licensed core service without a valid license")
	}
	if !config.ServiceConfigured(spec) {
		m.mu.Unlock()
		return fmt.Errorf("service %q is not configured: %s", name, config.ServiceConfigurationError(spec))
	}
	if spec.Kind == "static-assets" {
		status := m.statuses[name]
		status.State = "ready"
		status.LastError = ""
		m.statuses[name] = status
		m.mu.Unlock()
		return nil
	}
	if len(spec.Commands) == 0 {
		m.mu.Unlock()
		return fmt.Errorf("service %q commands are not configured", name)
	}
	if len(m.processes[name]) > 0 {
		m.mu.Unlock()
		return fmt.Errorf("service %q is already running", name)
	}

	started := make([]processEntry, 0, len(spec.Commands))
	status := buildInitialStatus(spec)
	status.State = "starting"
	status.Components = make([]ComponentStatus, 0, len(spec.Commands))

	for _, component := range spec.Commands {
		if err := ctx.Err(); err != nil {
			for _, startedEntry := range started {
				if startedEntry.command.Process != nil {
					_ = startedEntry.command.Process.Kill()
				}
			}
			m.mu.Unlock()
			return err
		}

		command := exec.Command(component.Executable, component.Args...)
		command.Env = mergeEnv(spec.Env)
		command.Dir = component.WorkingDir
		if err := command.Start(); err != nil {
			for _, startedEntry := range started {
				if startedEntry.command.Process != nil {
					_ = startedEntry.command.Process.Kill()
				}
			}
			status.State = "error"
			status.LastError = err.Error()
			status.Components = append(status.Components, ComponentStatus{
				Name:       component.Name,
				Executable: component.Executable,
				State:      "error",
				LastError:  err.Error(),
			})
			m.statuses[name] = status
			m.mu.Unlock()
			return err
		}

		started = append(started, processEntry{name: component.Name, command: command})
		status.Components = append(status.Components, ComponentStatus{
			Name:       component.Name,
			Executable: component.Executable,
			State:      "running",
			PID:        command.Process.Pid,
			StartedAt:  time.Now().UTC().Format(time.RFC3339),
		})
	}

	status.State = "running"
	m.processes[name] = started
	m.statuses[name] = status
	m.mu.Unlock()

	for _, entry := range started {
		go m.wait(name, entry)
	}
	return nil
}

func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	entries := append([]processEntry(nil), m.processes[name]...)
	if len(entries) == 0 {
		m.mu.Unlock()
		return fmt.Errorf("service %q is not running", name)
	}
	m.mu.Unlock()

	var stopErr error
	for _, entry := range entries {
		if entry.command.Process == nil {
			continue
		}
		if err := entry.command.Process.Signal(os.Interrupt); err != nil {
			if killErr := entry.command.Process.Kill(); killErr != nil {
				stopErr = errors.Join(stopErr, err, killErr)
			}
		}
	}
	return stopErr
}

func (m *Manager) Restart(ctx context.Context, name string, licenseValid bool) error {
	if status := m.statusByName(name); status.State == "running" {
		if err := m.Stop(name); err != nil {
			return err
		}
		time.Sleep(150 * time.Millisecond)
	}
	return m.Start(ctx, name, licenseValid)
}

func (m *Manager) wait(serviceName string, entry processEntry) {
	err := entry.command.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	entries := m.processes[serviceName]
	filtered := entries[:0]
	for _, item := range entries {
		if item.name != entry.name {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		delete(m.processes, serviceName)
	} else {
		m.processes[serviceName] = filtered
	}

	status := m.statuses[serviceName]
	for index, component := range status.Components {
		if component.Name != entry.name {
			continue
		}
		component.PID = 0
		component.StoppedAt = time.Now().UTC().Format(time.RFC3339)
		if err != nil {
			component.State = "error"
			component.LastError = exitError(err)
		} else {
			component.State = "stopped"
			component.LastError = ""
		}
		status.Components[index] = component
	}
	status.State = aggregateServiceState(status.Components, status.Kind)
	status.LastError = serviceLastError(status.Components)
	m.statuses[serviceName] = status
}

func (m *Manager) statusByName(name string) ServiceStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statuses[name]
}

func (m *Manager) orderedNames() []string {
	names := append([]string(nil), m.order...)
	sort.Strings(names)
	return names
}
