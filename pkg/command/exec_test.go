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
				assert.Nil(t, c.SysProcAttr)
				assert.Zero(t, c.WaitDelay)

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
				assert.Equal(t, processGroupWaitDelay, c.WaitDelay)
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

func Test_command_processGroupCancellationAllowsGracefulShutdown(t *testing.T) {
	v := validator.New()
	tempDir := t.TempDir()
	readyPath := filepath.Join(tempDir, "ready")
	terminatedPath := filepath.Join(tempDir, "terminated")
	cmd, err := New(v,
		WithCmd("sh"),
		WithArgs([]string{"-c", `trap 'printf terminated > "$TERMINATED_PATH"; exit 0' TERM; printf ready > "$READY_PATH"; while :; do sleep 1 & wait $!; done`}),
		WithEnv(map[string]string{
			"READY_PATH":      readyPath,
			"TERMINATED_PATH": terminatedPath,
		}),
		WithStdout(io.Discard),
		WithStderr(io.Discard),
		WithProcessGroup(),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- cmd.Exec(ctx)
	}()

	require.Eventually(t, func() bool {
		_, err := os.Stat(readyPath)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)

	cancel()
	select {
	case err := <-resultCh:
		require.Error(t, err)
	case <-time.After(processGroupTerminationGracePeriod + time.Second):
		t.Fatal("command did not return after cancellation")
	}

	require.Eventually(t, func() bool {
		_, err := os.Stat(terminatedPath)
		return err == nil
	}, time.Second, 10*time.Millisecond)
}

func Test_command_processGroupCancellationKillsStubbornChildren(t *testing.T) {
	v := validator.New()
	tempDir := t.TempDir()
	parentPIDPath := filepath.Join(tempDir, "parent.pid")
	childPIDPath := filepath.Join(tempDir, "child.pid")
	cmd, err := New(v,
		WithCmd("sh"),
		WithArgs([]string{"-c", `trap '' TERM; echo $$ > "$PARENT_PID_PATH"; sleep 30 & echo $! > "$CHILD_PID_PATH"; wait`}),
		WithEnv(map[string]string{
			"PARENT_PID_PATH": parentPIDPath,
			"CHILD_PID_PATH":  childPIDPath,
		}),
		WithStdout(io.Discard),
		WithStderr(io.Discard),
		WithProcessGroup(),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- cmd.Exec(ctx)
	}()

	parentPID := readPID(t, parentPIDPath)
	childPID := readPID(t, childPIDPath)
	processesRunning := true
	t.Cleanup(func() {
		if processesRunning {
			_ = syscall.Kill(parentPID, syscall.SIGKILL)
			_ = syscall.Kill(childPID, syscall.SIGKILL)
		}
	})

	cancelledAt := time.Now()
	cancel()
	select {
	case err := <-resultCh:
		require.Error(t, err)
	case <-time.After(processGroupWaitDelay + time.Second):
		t.Fatal("command did not return after forced cancellation")
	}
	require.GreaterOrEqual(t, time.Since(cancelledAt), processGroupTerminationGracePeriod)

	require.Eventually(t, func() bool {
		return processExited(parentPID) && processExited(childPID)
	}, time.Second, 10*time.Millisecond)
	processesRunning = false
}

func Test_command_processGroupWaitDelayBoundsInheritedPipes(t *testing.T) {
	v := validator.New()
	childPIDPath := filepath.Join(t.TempDir(), "child.pid")
	cmd, err := New(v,
		WithCmd("sh"),
		WithArgs([]string{"-c", `sleep 30 & echo $! > "$CHILD_PID_PATH"`}),
		WithEnv(map[string]string{"CHILD_PID_PATH": childPIDPath}),
		WithStdout(io.Discard),
		WithStderr(io.Discard),
		WithProcessGroup(),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), processGroupWaitDelay+time.Second)
	defer cancel()
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- cmd.Exec(ctx)
	}()

	childPID := readPID(t, childPIDPath)
	defer syscall.Kill(childPID, syscall.SIGKILL)

	select {
	case err := <-resultCh:
		require.ErrorIs(t, err, exec.ErrWaitDelay)
	case <-time.After(processGroupWaitDelay + time.Second):
		t.Fatal("command did not return after its inherited-pipe wait delay")
	}
}

func readPID(t *testing.T, path string) int {
	t.Helper()

	var pid int
	require.Eventually(t, func() bool {
		contents, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		pid, err = strconv.Atoi(strings.TrimSpace(string(contents)))
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	return pid
}

func processExited(pid int) bool {
	err := syscall.Kill(pid, 0)
	return errors.Is(err, syscall.ESRCH)
}
