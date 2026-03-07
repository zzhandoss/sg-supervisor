package servicehost

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(context.Context, Action) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, action Action) error {
	command := exec.CommandContext(ctx, action.Command, action.Args...)
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("%s failed: %w", action.Name, err)
	}
	return fmt.Errorf("%s failed: %w: %s", action.Name, err, message)
}

func ExecuteInstall(ctx context.Context, plan Plan, runner Runner) error {
	actions, err := InstallActions(plan)
	if err != nil {
		return err
	}
	return executeActions(ctx, runner, actions)
}

func ExecuteRepair(ctx context.Context, plan Plan, runner Runner) error {
	actions, err := RepairActions(plan)
	if err != nil {
		return err
	}
	return executeActions(ctx, runner, actions)
}

func ExecuteUninstall(ctx context.Context, plan Plan, runner Runner) error {
	actions, err := UninstallActions(plan)
	if err != nil {
		return err
	}
	return executeActions(ctx, runner, actions)
}

func executeActions(ctx context.Context, runner Runner, actions []Action) error {
	for _, action := range actions {
		err := runner.Run(ctx, action)
		if err == nil || action.IgnoreFailure {
			continue
		}
		return err
	}
	return nil
}
