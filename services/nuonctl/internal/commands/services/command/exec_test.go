package command

import (
	"context"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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
					WithStdout(ioutil.Discard),
					WithStderr(ioutil.Discard),
					WithCwd("/tmp/test"),
				)
				assert.NoError(t, err)

				return cmd
			},
			assertFn: func(t *testing.T, c *exec.Cmd) {
				assert.Equal(t, "/tmp/test/ls", c.Path)
				assert.Equal(t, "-alh", c.Args[1])
				assert.Equal(t, ioutil.Discard, c.Stdout)
				assert.Equal(t, ioutil.Discard, c.Stderr)
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			cmd := test.commandFn(t)

			execCmd := cmd.buildCommand(ctx)
			assert.NotNil(t, execCmd)
			test.assertFn(t, execCmd)
		})
	}

}
