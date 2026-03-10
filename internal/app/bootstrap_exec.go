package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runBootstrapCommand(ctx context.Context, dir string, env map[string]string, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	command.Env = os.Environ()
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	var buffer bytes.Buffer
	command.Stdout = &buffer
	command.Stderr = &buffer
	if err := command.Run(); err != nil {
		output := strings.TrimSpace(buffer.String())
		if output == "" {
			return "", fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
		}
		return output, fmt.Errorf("%s %s failed: %s", name, strings.Join(args, " "), output)
	}
	return strings.TrimSpace(buffer.String()), nil
}
