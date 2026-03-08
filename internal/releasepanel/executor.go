package releasepanel

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type CommandError struct {
	Err    error
	Output string
}

func (e *CommandError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Output) != "" {
		return strings.TrimSpace(e.Output)
	}
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

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
	if err != nil {
		return buffer.Bytes(), &CommandError{Err: err, Output: buffer.String()}
	}
	return buffer.Bytes(), nil
}

func commandOutput(err error) string {
	var commandErr *CommandError
	if errors.As(err, &commandErr) {
		return strings.TrimSpace(commandErr.Output)
	}
	return ""
}
