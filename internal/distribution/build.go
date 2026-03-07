package distribution

import "sg-supervisor/internal/packaging"

func Build(root string, stage packaging.AssembleReport) (Report, error) {
	switch stage.Platform {
	case "linux":
		return buildLinux(root, stage)
	case "windows":
		return buildWindows(root, stage)
	default:
		return Report{}, ErrUnsupportedPlatform
	}
}
