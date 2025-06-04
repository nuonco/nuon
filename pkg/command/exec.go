package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/abiosoft/lineprefix"
	"github.com/pkg/errors"
)

func (c *command) ExecWithOutput(ctx context.Context) ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("must set stdout to nil for output")
	}

	cmd, err := c.buildCommand(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build command")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unable to get command output: %w", err)
	}

	return output, nil
}

func (c *command) Exec(ctx context.Context) error {
	cmd, err := c.buildCommand(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to build command")
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to exec command: %w", err)
	}

	return nil
}

//nolint:gosec
func (c *command) buildCommand(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)

	envVars := os.Environ()
	for k, v := range c.Env {
		envVars = append(envVars, k+"="+v)
	}

	// create the correct stdout/stderr handlers
	// TODO(jm): pull this into it's own function
	stdout := c.Stdout
	stderr := c.Stderr
	opts := make([]lineprefix.Option, 0)

	if c.LinePrefixFn != nil {
		if c.LineColor != nil {
			opts = append(opts, lineprefix.PrefixFunc(func() string {
				return c.LineColor.Sprint(c.LinePrefixFn())
			}))
		} else {
			opts = append(opts, lineprefix.PrefixFunc(c.LinePrefixFn))
		}
	}
	if c.LinePrefix != "" {
		prefix := c.LinePrefix
		if c.LineColor != nil {
			prefix = c.LineColor.Sprintf("%s", prefix)
		}

		opts = append(opts, lineprefix.Prefix(prefix))
	}
	if c.LineColor != nil {
		opts = append(opts, lineprefix.Color(c.LineColor))
	}
	if len(opts) > 0 {
		stdout = lineprefix.New(opts...)
		stderr = lineprefix.New(opts...)
	}

	// if file output path is set, we also write to that.
	if c.FileOutputPath != "" {
		fpWriter, err := os.OpenFile(c.FileOutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open file output path")
		}

		stdout = io.MultiWriter(stdout, fpWriter)
		stderr = io.MultiWriter(stderr, fpWriter)
	}

	// build the command
	cmd.Env = envVars
	cmd.Stdin = c.Stdin
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Dir = c.Cwd
	return cmd, nil
}
