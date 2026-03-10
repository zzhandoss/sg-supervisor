package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ApplyRuntimeConfig(layout Layout, catalog ServiceCatalog, product ProductConfig, internal InternalRuntimeConfig) ServiceCatalog {
	status := deriveProductConfigStatus(product, detectMachineHosts())
	services := make([]ServiceSpec, 0, len(catalog.Services))
	for _, service := range catalog.Services {
		current := service
		current.Env = cloneEnv(service.Env)
		current.Env["SG_PRODUCT_ENV_FILE"] = ProductEnvFile(layout)
		current.Env["SG_INTERNAL_ENV_FILE"] = InternalRuntimeEnvFile(layout)
		current.Env["API_CORS_ALLOWED_ORIGINS"] = strings.Join(status.APICorsAllowedOrigins, ",")
		current.Env["DEVICE_SERVICE_CORS_ALLOWED_ORIGINS"] = strings.Join(status.DeviceServiceCorsAllowedOrigins, ",")
		injectInternalRuntimeEnv(current.Name, current.Env, layout, product, status, internal)
		services = append(services, current)
	}
	return ServiceCatalog{Services: services}
}

func injectInternalRuntimeEnv(serviceName string, env map[string]string, layout Layout, product ProductConfig, status ProductConfigStatus, internal InternalRuntimeConfig) {
	coreDB := filepath.Join(layout.DataDir, "school-gate", "app.db")
	deviceDB := filepath.Join(layout.DataDir, "school-gate", "device.db")
	apiURL := "http://127.0.0.1:3000"
	deviceServiceURL := "http://127.0.0.1:4010"
	botURL := "http://127.0.0.1:4100"

	env["DB_FILE"] = coreDB
	env["DEVICE_DB_FILE"] = deviceDB
	env["CORE_TOKEN"] = internal.CoreToken
	env["CORE_HMAC_SECRET"] = internal.CoreHMACSecret
	env["ADMIN_JWT_SECRET"] = internal.AdminJWTSecret
	env["DEVICE_SERVICE_TOKEN"] = internal.DeviceServiceToken
	env["DEVICE_SERVICE_INTERNAL_TOKEN"] = internal.DeviceServiceInternalKey
	env["BOT_INTERNAL_TOKEN"] = internal.BotInternalToken
	env["CORE_BASE_URL"] = apiURL
	env["MONITORING_API_URL"] = apiURL
	env["MONITORING_DEVICE_SERVICE_URL"] = deviceServiceURL
	env["MONITORING_BOT_URL"] = botURL

	if serviceName == "bot" {
		delete(env, "TELEGRAM_BOT_TOKEN")
		if product.TelegramBotToken != "" {
			env["TELEGRAM_BOT_TOKEN"] = product.TelegramBotToken
		}
	}
	if serviceName == "admin-ui" {
		env["VITE_API_BASE_URL"] = status.ViteAPIBaseURL
		env["HOST"] = "0.0.0.0"
		env["PORT"] = "5000"
		env["NITRO_HOST"] = "0.0.0.0"
		env["NITRO_PORT"] = "5000"
	}
	if serviceName == "dahua-terminal-adapter" {
		env["BASE_URL"] = "http://127.0.0.1:8091"
		env["ADAPTER_INSTANCE_KEY"] = "dahua-adapter"
		env["ADAPTER_INSTANCE_NAME"] = "dahua-adapter"
		env["DS_BASE_URL"] = deviceServiceURL
		env["DS_BEARER_TOKEN"] = internal.DeviceServiceToken
		env["DS_HMAC_SECRET"] = internal.CoreHMACSecret
		env["BACKFILL_BEARER_TOKEN"] = internal.DeviceServiceToken
		env["PUSH_DIGEST_REALM"] = "dahua-adapter"
	}
}

func ServiceConfigured(service ServiceSpec) bool {
	if service.Kind == "static-assets" {
		return strings.TrimSpace(service.StaticDir) != ""
	}
	if len(service.Commands) == 0 {
		return false
	}
	for _, key := range requiredServiceEnv(service.Name) {
		if strings.TrimSpace(service.Env[key]) == "" {
			return false
		}
	}
	return true
}

func ServiceConfigurationError(service ServiceSpec) string {
	missing := make([]string, 0, len(requiredServiceEnv(service.Name)))
	for _, key := range requiredServiceEnv(service.Name) {
		if strings.TrimSpace(service.Env[key]) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("missing required configuration: %s", strings.Join(missing, ", "))
}

func requiredServiceEnv(name string) []string {
	switch name {
	case "api":
		return []string{"CORE_TOKEN", "CORE_HMAC_SECRET", "ADMIN_JWT_SECRET", "DB_FILE"}
	case "device-service":
		return []string{"DEVICE_SERVICE_TOKEN", "DEVICE_SERVICE_INTERNAL_TOKEN", "CORE_BASE_URL", "CORE_TOKEN", "CORE_HMAC_SECRET", "ADMIN_JWT_SECRET", "DEVICE_DB_FILE"}
	case "bot":
		return []string{"BOT_INTERNAL_TOKEN", "TELEGRAM_BOT_TOKEN"}
	case "worker":
		return []string{"BOT_INTERNAL_TOKEN", "DEVICE_SERVICE_INTERNAL_TOKEN", "DB_FILE"}
	case "dahua-terminal-adapter":
		return []string{"BASE_URL", "ADAPTER_INSTANCE_KEY", "ADAPTER_INSTANCE_NAME", "DS_BASE_URL", "DS_BEARER_TOKEN", "DS_HMAC_SECRET", "BACKFILL_BEARER_TOKEN"}
	default:
		return nil
	}
}
