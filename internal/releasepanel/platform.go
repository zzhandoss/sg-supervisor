package releasepanel

import (
	"errors"
	"runtime"
)

var errUnsupportedHostPlatform = errors.New("local release is supported only on windows and linux hosts")

func hostPlatform() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows", nil
	case "linux":
		return "linux", nil
	default:
		return "", errUnsupportedHostPlatform
	}
}
