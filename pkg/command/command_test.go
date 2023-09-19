package command

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	cmd := generics.GetFakeObj[string]()
	args := generics.GetFakeObj[[]string]()
	env := generics.GetFakeObj[map[string]string]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []commandOption
		assertFn    func(*testing.T, *command)
	}{
		"happy path": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithArgs(args),
					WithEnv(env),
					WithStderr(io.Discard),
					WithStdout(io.Discard),
				}
			},
			assertFn: func(t *testing.T, l *command) {
				assert.Equal(t, cmd, l.Cmd)
				assert.Equal(t, args, l.Args)
				assert.Equal(t, env, l.Env)
				assert.Equal(t, io.Discard, l.Stderr)
				assert.Equal(t, io.Discard, l.Stdout)
			},
		},
		"uses stdout/err/in by default": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithArgs(args),
					WithEnv(env),
				}
			},
			assertFn: func(t *testing.T, l *command) {
				assert.Equal(t, os.Stderr, l.Stderr)
				assert.Equal(t, os.Stdin, l.Stdin)
				assert.Equal(t, os.Stdout, l.Stdout)
			},
		},
		"sets cwd": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithArgs(args),
					WithEnv(env),
					WithCwd("/home/path"),
				}
			},
			assertFn: func(t *testing.T, l *command) {
				assert.Equal(t, l.Cwd, "/home/path")
			},
		},
		"inherits env from local": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithArgs(args),
					WithInheritedEnv(),
				}
			},
			assertFn: func(t *testing.T, l *command) {
				assert.NotEmpty(t, l.Env)
			},
		},
		"missing cmd": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithEnv(env),
					WithArgs(args),
				}
			},
			errExpected: fmt.Errorf("Cmd"),
		},
		"missing args": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithEnv(env),
				}
			},
			errExpected: fmt.Errorf("Args"),
		},
		"missing env": {
			optsFn: func() []commandOption {
				return []commandOption{
					WithCmd(cmd),
					WithArgs(args),
				}
			},
			errExpected: fmt.Errorf("Env"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := New(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, e)
		})
	}
}
