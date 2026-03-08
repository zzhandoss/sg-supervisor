package distribution

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

func mountWindowsBuildStage(sourceDir string) (string, func(), error) {
	substPath, err := exec.LookPath("subst.exe")
	if err != nil {
		substPath, err = exec.LookPath("subst")
		if err != nil {
			return "", nil, err
		}
	}
	absoluteSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return "", nil, err
	}
	letters := []string{"X:", "Y:", "Z:", "W:", "V:", "U:", "T:", "S:", "R:", "Q:", "P:"}
	var lastErr error
	for _, letter := range letters {
		command := exec.Command(substPath, letter, absoluteSourceDir)
		output, err := command.CombinedOutput()
		if err == nil {
			return letter + `\`, func() {
				_ = exec.Command(substPath, letter, "/D").Run()
			}, nil
		}
		lastErr = errors.New(strings.TrimSpace(string(output)))
		if lastErr.Error() == "" {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no free drive letters for subst")
	}
	return "", nil, lastErr
}
