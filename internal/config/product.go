package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type ProductConfig struct {
	TelegramBotToken string `json:"telegramBotToken,omitempty"`
}

type ProductStore struct {
	configPath string
	envPath    string
}

func NewProductStore(layout Layout) *ProductStore {
	return &ProductStore{
		configPath: filepath.Join(layout.ConfigDir, "product-config.json"),
		envPath:    filepath.Join(layout.RuntimeDir, "config", "product.env"),
	}
}

func ProductEnvFile(layout Layout) string {
	return filepath.Join(layout.RuntimeDir, "config", "product.env")
}

func (s *ProductStore) Ensure() error {
	if err := os.MkdirAll(filepath.Dir(s.envPath), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return s.Save(ProductConfig{})
	} else if err != nil {
		return err
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return err
	}
	var cfg ProductConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	return writeProductEnvFile(s.envPath, cfg)
}

func (s *ProductStore) Load() (ProductConfig, error) {
	if err := s.Ensure(); err != nil {
		return ProductConfig{}, err
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return ProductConfig{}, err
	}
	var cfg ProductConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ProductConfig{}, err
	}
	return cfg, nil
}

func (s *ProductStore) Save(cfg ProductConfig) error {
	cfg.TelegramBotToken = strings.TrimSpace(cfg.TelegramBotToken)
	if strings.ContainsAny(cfg.TelegramBotToken, "\r\n") {
		return errors.New("telegram bot token must not contain newlines")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.configPath, data, 0o644); err != nil {
		return err
	}
	return writeProductEnvFile(s.envPath, cfg)
}

func (s *ProductStore) SetTelegramBotToken(token string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	cfg.TelegramBotToken = token
	return s.Save(cfg)
}

func writeProductEnvFile(path string, cfg ProductConfig) error {
	lines := make([]string, 0, 1)
	if cfg.TelegramBotToken != "" {
		lines = append(lines, "TELEGRAM_BOT_TOKEN="+cfg.TelegramBotToken)
	}
	data := strings.Join(lines, "\n")
	if data != "" {
		data += "\n"
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
