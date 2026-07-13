package command

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_command_buildCommand(t *testing.T) {
	v := validator.New()

	tests := map[string]struct {
		commandFn func(*testing.T) *command
		assertFn  func(*testing.T, *exec.Cmd)
	}{
		"happy path": {
			commandFn: func(t *testing.T) *command {
				cmd, err := New(v, WithCmd("/tmp/test/ls"),
					WithArgs([]string{"-alh"}),
					WithEnv(map[string]string{"KEY": "VALUE"}),
					WithStdout(io.Discard),
					WithStderr(io.Discard),
					WithCwd("/tmp/test"),
				)
				assert.NoError(t, err)

				return cmd
			},
			assertFn: func(t *testing.T, c *exec.Cmd) {
				assert.Equal(t, "/tmp/test/ls", c.Path)
				assert.Equal(t, "-alh", c.Args[1])
				assert.Equal(t, io.Discard, c.Stdout)
				assert.Equal(t, io.Discard, c.Stderr)
				assert.Equal(t, "/tmp/test", c.Dir)

				found := false
				for _, kv := range c.Env {
					if kv == "KEY=VALUE" {
						found = true
					}
				}
				assert.True(t, found)
			},
		},
		"process group": {
			commandFn: func(t *testing.T) *command {
				cmd, err := New(v,
					WithCmd("/tmp/test/ls"),
					WithEnv(map[string]string{}),
					WithProcessGroup(),
				)
				assert.NoError(t, err)

				return cmd
			},
			assertFn: func(t *testing.T, c *exec.Cmd) {
				require.NotNil(t, c.SysProcAttr)
				assert.True(t, c.SysProcAttr.Setpgid)
				assert.NotNil(t, c.Cancel)
				assert.Equal(t, 3*time.Second, c.WaitDelay)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			cmd := test.commandFn(t)

			execCmd, _, err := cmd.buildCommand(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, execCmd)
			test.assertFn(t, execCmd)
		})
	}
}

func Test_command_processGroupCancellation(t *testing.T) {
	v := validator.New()
	childPIDPath := filepath.Join(t.TempDir(), "child.pid")
	cmd, err := New(v,
		WithCmd("sh"),
		WithArgs([]string{"-c", `sleep 30 & echo $! > "$CHILD_PID_PATH"; wait`}),
		WithEnv(map[string]string{"CHILD_PID_PATH": childPIDPath}),
		WithStdout(io.Discard),
		WithStderr(io.Discard),
		WithProcessGroup(),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- cmd.Exec(ctx)
	}()

	var childPID int
	require.Eventually(t, func() bool {
		contents, err := os.ReadFile(childPIDPath)
		if err != nil {
			return false
		}
		childPID, err = strconv.Atoi(strings.TrimSpace(string(contents)))
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	childRunning := true
	t.Cleanup(func() {
		if childRunning {
			_ = syscall.Kill(childPID, syscall.SIGKILL)
		}
	})

	cancel()
	select {
	case err := <-resultCh:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("command did not return after cancellation")
	}

	require.Eventually(t, func() bool {
		err := syscall.Kill(childPID, 0)
		return errors.Is(err, syscall.ESRCH)
	}, 5*time.Second, 10*time.Millisecond)
	childRunning = false
}
