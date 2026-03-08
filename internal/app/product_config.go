package app

import (
	"context"
	"strings"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/control"
	"sg-supervisor/internal/setup"
)

func (a *App) ProductConfigStatus(ctx context.Context) (control.ProductConfigStatus, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.ProductConfigStatus{}, err
	}
	status, err := a.product.Status()
	if err != nil {
		return control.ProductConfigStatus{}, err
	}
	return mapProductConfigStatus(status), nil
}

func (a *App) UpdateProductConfig(ctx context.Context, update control.ProductConfigUpdate) (control.ProductConfigStatus, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.ProductConfigStatus{}, err
	}
	cfg, err := a.product.Load()
	if err != nil {
		return control.ProductConfigStatus{}, err
	}
	if update.PreferredHost != nil {
		cfg.PreferredHost = strings.TrimSpace(*update.PreferredHost)
	}
	if update.ClearTelegramBotToken {
		cfg.TelegramBotToken = ""
	}
	if update.TelegramBotToken != nil {
		cfg.TelegramBotToken = strings.TrimSpace(*update.TelegramBotToken)
	}
	if err := a.product.Save(cfg); err != nil {
		return control.ProductConfigStatus{}, err
	}
	if err := a.syncProductSetupState(ctx, cfg); err != nil {
		return control.ProductConfigStatus{}, err
	}
	if err := a.syncRuntimeConfig(); err != nil {
		return control.ProductConfigStatus{}, err
	}
	return a.ProductConfigStatus(ctx)
}

func (a *App) syncProductSetupState(ctx context.Context, cfg config.ProductConfig) error {
	status := setup.StatusPending
	if strings.TrimSpace(cfg.TelegramBotToken) != "" {
		status = setup.StatusCompleted
	}
	_, err := a.setup.UpdateField(ctx, setup.FieldTelegramBot, status)
	return err
}

func mapProductConfigStatus(status config.ProductConfigStatus) control.ProductConfigStatus {
	return control.ProductConfigStatus{
		PreferredHost:                   status.PreferredHost,
		ResolvedHost:                    status.ResolvedHost,
		AvailableHosts:                  append([]string(nil), status.AvailableHosts...),
		TelegramBotConfigured:           status.TelegramBotConfigured,
		ViteAPIBaseURL:                  status.ViteAPIBaseURL,
		AdminUIURL:                      status.AdminUIURL,
		APICorsAllowedOrigins:           append([]string(nil), status.APICorsAllowedOrigins...),
		DeviceServiceCorsAllowedOrigins: append([]string(nil), status.DeviceServiceCorsAllowedOrigins...),
	}
}
