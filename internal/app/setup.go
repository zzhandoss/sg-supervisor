package app

import (
	"context"

	"sg-supervisor/internal/control"
	"sg-supervisor/internal/setup"
)

func (a *App) SetupStatus(ctx context.Context) (control.SetupStatus, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.SetupStatus{}, err
	}
	licenseStatus, err := a.license.Status(ctx)
	if err != nil {
		return control.SetupStatus{}, err
	}
	state, err := a.setup.Load(ctx)
	if err != nil {
		return control.SetupStatus{}, err
	}
	return mapSetupStatus(setup.Summarize(state, licenseStatus.Valid)), nil
}

func (a *App) UpdateSetupField(ctx context.Context, key, status, value string) (control.SetupStatus, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.SetupStatus{}, err
	}
	if err := a.applySetupFieldValue(key, status, value); err != nil {
		return control.SetupStatus{}, err
	}
	state, err := a.setup.UpdateField(ctx, key, status)
	if err != nil {
		return control.SetupStatus{}, err
	}
	if err := a.syncRuntimeConfig(); err != nil {
		return control.SetupStatus{}, err
	}
	licenseStatus, err := a.license.Status(ctx)
	if err != nil {
		return control.SetupStatus{}, err
	}
	return mapSetupStatus(setup.Summarize(state, licenseStatus.Valid)), nil
}

func mapSetupStatus(summary setup.Summary) control.SetupStatus {
	return control.SetupStatus{
		Complete:       summary.Complete,
		BlockingFields: append([]string(nil), summary.BlockingFields...),
		Required:       mapSetupFields(summary.Required),
		Optional:       mapSetupFields(summary.Optional),
	}
}

func mapSetupFields(fields []setup.Field) []control.SetupField {
	result := make([]control.SetupField, 0, len(fields))
	for _, field := range fields {
		result = append(result, control.SetupField{
			Key:      field.Key,
			Label:    field.Label,
			Required: field.Required,
			Status:   field.Status,
		})
	}
	return result
}
