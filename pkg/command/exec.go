package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func (c *command) ExecWithOutput(ctx context.Context) ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("must set stdout to nil for output")
	}

	cmd := c.buildCommand(ctx)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unable to get command output: %w", err)
	}

	return output, nil
}

func (c *command) Exec(ctx context.Context) error {
	cmd := c.buildCommand(ctx)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}

	return nil
}

//nolint:gosec
func (c *command) buildCommand(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)

	envVars := os.Environ()
	for k, v := range c.Env {
		envVars = append(envVars, k+"="+v)
	}

	cmd.Env = envVars
	cmd.Stdin = c.Stdin
	cmd.Stderr = c.Stderr
	cmd.Stdout = c.Stdout
	cmd.Dir = c.Cwd
	return cmd
}
