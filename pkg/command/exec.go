package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/abiosoft/lineprefix"
	"github.com/pkg/errors"
)

func (c *command) ExecWithOutput(ctx context.Context) ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("must set stdout to nil for output")
	}

	cmd, cleanup, err := c.buildCommand(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build command")
	}
	defer cleanup()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unable to get command output: %w", err)
	}

	return output, nil
}

func (c *command) Exec(ctx context.Context) error {
	cmd, cleanup, err := c.buildCommand(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to build command")
	}
	defer cleanup()

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to exec command: %w", err)
	}

	return nil
}

//nolint:gosec
func (c *command) buildCommand(ctx context.Context) (*exec.Cmd, func(), error) {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)

	if c.UseProcessGroup {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			pgid := -cmd.Process.Pid
			// Send SIGTERM first so children can clean up
			_ = syscall.Kill(pgid, syscall.SIGTERM)

			// Give processes time to exit gracefully, then force kill
			done := make(chan struct{})
			go func() {
				cmd.Process.Wait()
				close(done)
			}()
			select {
			case <-done:
				return nil
			case <-time.After(3 * time.Second):
				return syscall.Kill(pgid, syscall.SIGKILL)
			}
		}
	}

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
	if quietMode {
		stderr = io.Discard
		stdout = io.Discard
	}

	// cleanup function to close any file handles
	cleanup := func() {}

	// if file output path is set, we also write to that.
	if c.FileOutputPath != "" {
		// Rotate existing log file before creating a new one
		if _, err := os.Stat(c.FileOutputPath); err == nil {
			rotatedPath := c.FileOutputPath + "." + time.Now().Format("2006-01-02T15-04-05")
			// Best-effort rename; if it fails, we'll just truncate
			os.Rename(c.FileOutputPath, rotatedPath)
		}

		fpWriter, err := os.OpenFile(c.FileOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to open file output path")
		}

		cleanup = func() {
			fpWriter.Close()
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

	return cmd, cleanup, nil
}
