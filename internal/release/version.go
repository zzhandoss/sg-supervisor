package release

import "strings"

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")
}
