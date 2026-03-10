package releasepanel

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type deliveryRecipe struct {
	InstallerVersion  string `json:"installerVersion"`
	SchoolGateVersion string `json:"schoolGateVersion"`
	AdapterVersion    string `json:"adapterVersion"`
	NodeVersion       string `json:"nodeVersion"`
}

func prepareDeliveryRoot(root string, state State, assets WorkspaceAssets, bundledFreeLicensePath string) (string, error) {
	payloadDir := filepath.Join(root, "payload")
	if err := os.MkdirAll(payloadDir, 0o755); err != nil {
		return "", err
	}
	if err := copyArtifact(assets.SchoolGateSourcePath, filepath.Join(payloadDir, filepath.Base(assets.SchoolGateSourcePath))); err != nil {
		return "", err
	}
	if err := copyArtifact(assets.AdapterPath, filepath.Join(payloadDir, filepath.Base(assets.AdapterPath))); err != nil {
		return "", err
	}
	if err := writeDeliveryRecipe(root, state.Recipe); err != nil {
		return "", err
	}
	if _, err := copyBundledFreeLicense(root, bundledFreeLicensePath); err != nil {
		return "", err
	}
	return root, nil
}

func writeDeliveryRecipe(root string, recipe Recipe) error {
	body := deliveryRecipe{
		InstallerVersion:  trimVersion(recipe.InstallerVersion),
		SchoolGateVersion: trimVersion(recipe.SchoolGateVersion),
		AdapterVersion:    trimVersion(recipe.AdapterVersion),
		NodeVersion:       trimVersion(recipe.NodeVersion),
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(root, "delivery-recipe.json"), data, 0o644)
}
