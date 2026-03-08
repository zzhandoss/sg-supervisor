package config

import (
	"fmt"
	"net"
	"slices"
	"strings"
)

var corsPorts = []int{3000, 5000, 4100}

func deriveProductConfigStatus(cfg ProductConfig, availableHosts []string) ProductConfigStatus {
	hosts := uniqueHosts(availableHosts)
	resolvedHost := resolvePreferredHost(cfg.PreferredHost, hosts)
	origins := buildOrigins(cfg.PreferredHost, hosts)

	return ProductConfigStatus{
		PreferredHost:                   cfg.PreferredHost,
		ResolvedHost:                    resolvedHost,
		AvailableHosts:                  hosts,
		TelegramBotConfigured:           strings.TrimSpace(cfg.TelegramBotToken) != "",
		ViteAPIBaseURL:                  fmt.Sprintf("http://%s:3000", resolvedHost),
		AdminUIURL:                      fmt.Sprintf("http://%s:5000", resolvedHost),
		APICorsAllowedOrigins:           origins,
		DeviceServiceCorsAllowedOrigins: append([]string(nil), origins...),
	}
}

func validatePreferredHost(host string) error {
	if host == "" {
		return nil
	}
	if strings.ContainsAny(host, "\r\n/\\") || strings.Contains(host, "://") {
		return fmt.Errorf("preferred host must be a host or ip without scheme or path")
	}
	return nil
}

func detectMachineHosts() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return []string{"localhost", "127.0.0.1"}
	}

	hosts := []string{"localhost", "127.0.0.1"}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil || ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			hosts = append(hosts, ip.String())
		}
	}
	return uniqueHosts(hosts)
}

func resolvePreferredHost(preferredHost string, availableHosts []string) string {
	if preferredHost != "" {
		return preferredHost
	}
	for _, host := range availableHosts {
		if host != "localhost" && host != "127.0.0.1" {
			return host
		}
	}
	return "127.0.0.1"
}

func buildOrigins(preferredHost string, availableHosts []string) []string {
	hosts := uniqueHosts(append(availableHosts, preferredHost))
	origins := make([]string, 0, len(hosts)*len(corsPorts))
	for _, host := range hosts {
		for _, port := range corsPorts {
			origins = append(origins, fmt.Sprintf("http://%s:%d", host, port))
		}
	}
	return origins
}

func uniqueHosts(hosts []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(hosts))
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		result = append(result, host)
	}
	slices.Sort(result)
	return result
}
