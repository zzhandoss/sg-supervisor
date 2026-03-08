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
	PreferredHost    string `json:"preferredHost,omitempty"`
}

type ProductConfigStatus struct {
	PreferredHost                   string   `json:"preferredHost,omitempty"`
	ResolvedHost                    string   `json:"resolvedHost"`
	AvailableHosts                  []string `json:"availableHosts"`
	TelegramBotConfigured           bool     `json:"telegramBotConfigured"`
	ViteAPIBaseURL                  string   `json:"viteApiBaseUrl"`
	AdminUIURL                      string   `json:"adminUiUrl"`
	APICorsAllowedOrigins           []string `json:"apiCorsAllowedOrigins"`
	DeviceServiceCorsAllowedOrigins []string `json:"deviceServiceCorsAllowedOrigins"`
}

type ProductStore struct {
	configPath string
	envPath    string
	hosts      func() []string
}

func NewProductStore(layout Layout) *ProductStore {
	return &ProductStore{
		configPath: filepath.Join(layout.ConfigDir, "product-config.json"),
		envPath:    filepath.Join(layout.RuntimeDir, "config", "product.env"),
		hosts:      detectMachineHosts,
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
	return writeProductEnvFile(s.envPath, cfg, s.hosts())
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
	cfg.PreferredHost = strings.TrimSpace(cfg.PreferredHost)
	if strings.ContainsAny(cfg.TelegramBotToken, "\r\n") {
		return errors.New("telegram bot token must not contain newlines")
	}
	if err := validatePreferredHost(cfg.PreferredHost); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.configPath, data, 0o644); err != nil {
		return err
	}
	return writeProductEnvFile(s.envPath, cfg, s.hosts())
}

func (s *ProductStore) SetTelegramBotToken(token string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	cfg.TelegramBotToken = token
	return s.Save(cfg)
}

func (s *ProductStore) Status() (ProductConfigStatus, error) {
	cfg, err := s.Load()
	if err != nil {
		return ProductConfigStatus{}, err
	}
	return deriveProductConfigStatus(cfg, s.hosts()), nil
}

func writeProductEnvFile(path string, cfg ProductConfig, availableHosts []string) error {
	status := deriveProductConfigStatus(cfg, availableHosts)
	lines := make([]string, 0, 4)
	if cfg.TelegramBotToken != "" {
		lines = append(lines, "TELEGRAM_BOT_TOKEN="+cfg.TelegramBotToken)
	}
	lines = append(lines, "VITE_API_BASE_URL="+status.ViteAPIBaseURL)
	lines = append(lines, "API_CORS_ALLOWED_ORIGINS="+strings.Join(status.APICorsAllowedOrigins, ","))
	lines = append(lines, "DEVICE_SERVICE_CORS_ALLOWED_ORIGINS="+strings.Join(status.DeviceServiceCorsAllowedOrigins, ","))
	data := strings.Join(lines, "\n")
	if data != "" {
		data += "\n"
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
