package app

import (
	"errors"
	"strings"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/setup"
)

func (a *App) syncRuntimeConfig() error {
	productCfg, err := a.product.Load()
	if err != nil {
		return err
	}
	internalCfg, err := a.internal.Load()
	if err != nil {
		return err
	}
	catalog, err := config.LoadServiceCatalog(a.layout)
	if err != nil {
		return err
	}
	a.runtime.Reconfigure(config.ApplyRuntimeConfig(a.layout, catalog, productCfg, internalCfg))
	return nil
}

func (a *App) applySetupFieldValue(key, status, value string) error {
	value = strings.TrimSpace(value)
	if key != setup.FieldTelegramBot {
		if value != "" {
			return errors.New("setup field does not accept a value")
		}
		return nil
	}
	if status == setup.StatusCompleted {
		if value == "" {
			cfg, err := a.product.Load()
			if err != nil {
				return err
			}
			if strings.TrimSpace(cfg.TelegramBotToken) == "" {
				return errors.New("telegram bot token is required when telegram-bot is completed")
			}
			return nil
		}
		return a.product.SetTelegramBotToken(value)
	}
	return a.product.SetTelegramBotToken("")
}
