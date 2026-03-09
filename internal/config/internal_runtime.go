package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type InternalRuntimeConfig struct {
	CoreToken                string `json:"coreToken"`
	CoreHMACSecret           string `json:"coreHmacSecret"`
	AdminJWTSecret           string `json:"adminJwtSecret"`
	DeviceServiceToken       string `json:"deviceServiceToken"`
	DeviceServiceInternalKey string `json:"deviceServiceInternalKey"`
	BotInternalToken         string `json:"botInternalToken"`
}

type InternalRuntimeStore struct {
	configPath string
	envPath    string
}

func NewInternalRuntimeStore(layout Layout) *InternalRuntimeStore {
	return &InternalRuntimeStore{
		configPath: filepath.Join(layout.ConfigDir, "internal-runtime.json"),
		envPath:    filepath.Join(layout.RuntimeDir, "config", "internal-runtime.env"),
	}
}

func InternalRuntimeEnvFile(layout Layout) string {
	return filepath.Join(layout.RuntimeDir, "config", "internal-runtime.env")
}

func (s *InternalRuntimeStore) Ensure() error {
	if err := os.MkdirAll(filepath.Dir(s.envPath), 0o755); err != nil {
		return err
	}
	cfg, changed, err := s.loadOrGenerate()
	if err != nil {
		return err
	}
	if changed {
		return s.Save(cfg)
	}
	return writeInternalRuntimeEnvFile(s.envPath, cfg)
}

func (s *InternalRuntimeStore) Load() (InternalRuntimeConfig, error) {
	if err := s.Ensure(); err != nil {
		return InternalRuntimeConfig{}, err
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return InternalRuntimeConfig{}, err
	}
	var cfg InternalRuntimeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return InternalRuntimeConfig{}, err
	}
	return cfg, nil
}

func (s *InternalRuntimeStore) Save(cfg InternalRuntimeConfig) error {
	cfg = normalizeInternalRuntimeConfig(cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.configPath, data, 0o644); err != nil {
		return err
	}
	return writeInternalRuntimeEnvFile(s.envPath, cfg)
}

func (s *InternalRuntimeStore) loadOrGenerate() (InternalRuntimeConfig, bool, error) {
	data, err := os.ReadFile(s.configPath)
	if os.IsNotExist(err) {
		return generateInternalRuntimeConfig(), true, nil
	}
	if err != nil {
		return InternalRuntimeConfig{}, false, err
	}
	var cfg InternalRuntimeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return InternalRuntimeConfig{}, false, err
	}
	normalized := normalizeInternalRuntimeConfig(cfg)
	return normalized, normalized != cfg, nil
}

func normalizeInternalRuntimeConfig(cfg InternalRuntimeConfig) InternalRuntimeConfig {
	cfg.CoreToken = ensureSecret(cfg.CoreToken, 24)
	cfg.CoreHMACSecret = ensureSecret(cfg.CoreHMACSecret, 32)
	cfg.AdminJWTSecret = ensureSecret(cfg.AdminJWTSecret, 32)
	cfg.DeviceServiceToken = ensureSecret(cfg.DeviceServiceToken, 24)
	cfg.DeviceServiceInternalKey = ensureSecret(cfg.DeviceServiceInternalKey, 24)
	cfg.BotInternalToken = ensureSecret(cfg.BotInternalToken, 24)
	return cfg
}

func generateInternalRuntimeConfig() InternalRuntimeConfig {
	return normalizeInternalRuntimeConfig(InternalRuntimeConfig{})
}

func ensureSecret(value string, size int) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func writeInternalRuntimeEnvFile(path string, cfg InternalRuntimeConfig) error {
	lines := []string{
		"CORE_TOKEN=" + cfg.CoreToken,
		"CORE_HMAC_SECRET=" + cfg.CoreHMACSecret,
		"ADMIN_JWT_SECRET=" + cfg.AdminJWTSecret,
		"DEVICE_SERVICE_TOKEN=" + cfg.DeviceServiceToken,
		"DEVICE_SERVICE_INTERNAL_TOKEN=" + cfg.DeviceServiceInternalKey,
		"BOT_INTERNAL_TOKEN=" + cfg.BotInternalToken,
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
