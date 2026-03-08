package runtime

import "sg-supervisor/internal/config"

func (m *Manager) Reconfigure(catalog config.ServiceCatalog) {
	m.mu.Lock()
	defer m.mu.Unlock()

	order := make([]string, 0, len(catalog.Services))
	specs := make(map[string]config.ServiceSpec, len(catalog.Services))
	statuses := make(map[string]ServiceStatus, len(catalog.Services))
	processes := make(map[string][]processEntry, len(catalog.Services))

	for _, service := range catalog.Services {
		order = append(order, service.Name)
		specs[service.Name] = service
		if current, ok := m.statuses[service.Name]; ok {
			statuses[service.Name] = mergeStatusWithSpec(current, service)
		} else {
			statuses[service.Name] = buildInitialStatus(service)
		}
		if entries, ok := m.processes[service.Name]; ok {
			processes[service.Name] = entries
		}
	}

	m.order = order
	m.specs = specs
	m.statuses = statuses
	m.processes = processes
}

func mergeStatusWithSpec(current ServiceStatus, service config.ServiceSpec) ServiceStatus {
	next := current
	next.Kind = service.Kind
	next.Configured = service.StaticDir != "" || len(service.Commands) > 0
	next.RequiresLicense = service.RequiresLicense
	next.StaticDir = service.StaticDir

	components := make([]ComponentStatus, 0, len(service.Commands))
	byName := make(map[string]ComponentStatus, len(current.Components))
	for _, component := range current.Components {
		byName[component.Name] = component
	}
	for _, command := range service.Commands {
		component, ok := byName[command.Name]
		if !ok {
			component = ComponentStatus{Name: command.Name, State: "stopped"}
		}
		component.Executable = command.Executable
		components = append(components, component)
	}
	next.Components = components
	return next
}
