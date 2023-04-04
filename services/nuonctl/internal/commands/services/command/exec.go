package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func (c *command) Exec(ctx context.Context) error {
	cmd := c.buildCommand(ctx)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error executing command: %w", err)
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
