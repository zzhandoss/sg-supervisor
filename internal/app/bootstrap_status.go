package app

import (
	"strings"
	"time"

	"sg-supervisor/internal/bootstrap"
)

func (a *App) markBootstrapStep(status *bootstrap.Status, name, state, message string) error {
	status.CurrentStep = name
	for index, step := range status.Steps {
		if step.Name != name {
			continue
		}
		step.State = state
		step.Message = message
		if state == "running" {
			step.StartedAt = time.Now().UTC().Format(time.RFC3339)
		}
		status.Steps[index] = step
		break
	}
	status.Logs = append(status.Logs, strings.TrimSpace(message))
	return a.bootstrap.Save(*status)
}

func (a *App) completeBootstrapStep(status *bootstrap.Status, name, message string) error {
	for index, step := range status.Steps {
		if step.Name != name {
			continue
		}
		step.State = "succeeded"
		step.Message = message
		step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		status.Steps[index] = step
		break
	}
	status.CurrentStep = ""
	status.Logs = append(status.Logs, strings.TrimSpace(message))
	return a.bootstrap.Save(*status)
}

func (a *App) failBootstrap(status bootstrap.Status, name string, err error) {
	status.State = "failed"
	status.CurrentStep = name
	status.Error = err.Error()
	status.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	status.Logs = append(status.Logs, err.Error())
	for index, step := range status.Steps {
		if step.Name != name {
			continue
		}
		step.State = "failed"
		step.Message = err.Error()
		step.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		status.Steps[index] = step
		break
	}
	_ = a.bootstrap.Save(status)
}
