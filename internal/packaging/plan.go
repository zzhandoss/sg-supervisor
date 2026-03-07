package packaging

import (
	"path/filepath"
	"strings"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/servicehost"
)

const Architecture = "x64"

type FileEntry struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	Kind       string `json:"kind"`
}

type Manifest struct {
	ProductName      string               `json:"productName"`
	Platform         string               `json:"platform"`
	Architecture     string               `json:"architecture"`
	PackageID        string               `json:"packageId,omitempty"`
	ActivePackageID  string               `json:"activePackageId,omitempty"`
	ListenAddress    string               `json:"listenAddress"`
	ServiceName      string               `json:"serviceName"`
	SupervisorBinary string               `json:"supervisorBinary"`
	Files            []FileEntry          `json:"files"`
	InstallActions   []servicehost.Action `json:"installActions"`
	UninstallActions []servicehost.Action `json:"uninstallActions"`
}

func BuildManifest(layout config.Layout, cfg config.SupervisorConfig, rendered servicehost.RenderedArtifacts, packageID, activePackageID, platform string) (Manifest, error) {
	installActions, err := servicehost.InstallActionsForTarget(rendered.Plan, platform)
	if err != nil {
		return Manifest{}, err
	}
	uninstallActions, err := servicehost.UninstallActionsForTarget(rendered.Plan, platform)
	if err != nil {
		return Manifest{}, err
	}

	files := []FileEntry{
		{SourcePath: rendered.Plan.BinaryPath, TargetPath: filepath.Join("supervisor", filepath.Base(rendered.Plan.BinaryPath)), Kind: "supervisor-binary"},
		{SourcePath: layout.InstallDir, TargetPath: "install", Kind: "product-install-root"},
	}
	for _, path := range serviceHostFilesForPlatform(rendered.WrittenFiles, platform) {
		files = append(files, FileEntry{
			SourcePath: path,
			TargetPath: filepath.Join("runtime", "service-host", platform, filepath.Base(path)),
			Kind:       "service-host-artifact",
		})
	}

	return Manifest{
		ProductName:      cfg.ProductName,
		Platform:         platform,
		Architecture:     Architecture,
		PackageID:        packageID,
		ActivePackageID:  activePackageID,
		ListenAddress:    rendered.Plan.ListenAddress,
		ServiceName:      rendered.Plan.ServiceName,
		SupervisorBinary: rendered.Plan.BinaryPath,
		Files:            files,
		InstallActions:   installActions,
		UninstallActions: uninstallActions,
	}, nil
}

func serviceHostFilesForPlatform(paths []string, platform string) []string {
	result := make([]string, 0, len(paths))
	needle := "/" + platform + "/"
	for _, path := range paths {
		if !strings.Contains(filepath.ToSlash(path), needle) {
			continue
		}
		result = append(result, path)
	}
	return result
}
