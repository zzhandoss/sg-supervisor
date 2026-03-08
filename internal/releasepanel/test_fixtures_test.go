package releasepanel

func coreFiles() map[string]string {
	return map[string]string{
		"school-gate-v1.2.0/apps/api/dist/index.js":                  "api",
		"school-gate-v1.2.0/apps/device-service/dist/api/main.js":    "device api",
		"school-gate-v1.2.0/apps/device-service/dist/outbox/main.js": "device outbox",
		"school-gate-v1.2.0/apps/bot/dist/main.js":                   "bot",
		"school-gate-v1.2.0/apps/worker/dist/main.js":                "worker",
		"school-gate-v1.2.0/apps/worker/dist/accessEvents/main.js":   "access",
		"school-gate-v1.2.0/apps/worker/dist/outbox/main.js":         "outbox",
		"school-gate-v1.2.0/apps/worker/dist/retention/main.js":      "retention",
		"school-gate-v1.2.0/apps/worker/dist/monitoring/main.js":     "monitoring",
		"school-gate-v1.2.0/apps/admin-ui/dist/index.html":           "admin",
	}
}

func adapterFiles() map[string]string {
	return map[string]string{
		"dahua-adapter-v0.2.0/dist/src/index.js": "adapter",
	}
}
