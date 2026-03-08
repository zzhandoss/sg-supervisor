package releasepanel

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

type Executor interface {
	Run(ctx context.Context, dir string, env map[string]string, name string, args ...string) ([]byte, error)
}

type ExecExecutor struct{}

func (ExecExecutor) Run(ctx context.Context, dir string, env map[string]string, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	command.Env = os.Environ()
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	var buffer bytes.Buffer
	command.Stdout = &buffer
	command.Stderr = &buffer
	err := command.Run()
	return buffer.Bytes(), err
}
