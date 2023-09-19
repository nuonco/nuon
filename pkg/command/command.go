package command

import (
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
)

type command struct {
	v *validator.Validate

	Cmd  string            `validate:"required"`
	Args []string          `validate:"required"`
	Env  map[string]string `validate:"required"`

	// non-optional arguments
	Cwd    string
	Stdout io.Writer
	Stdin  io.Reader `validate:"required"`
	Stderr io.Writer `validate:"required"`
}

type commandOption func(*command) error

func New(v *validator.Validate, opts ...commandOption) (*command, error) {
	l := &command{
		v:      v,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
	for idx, opt := range opts {
		if err := opt(l); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := l.v.Struct(l); err != nil {
		return nil, fmt.Errorf("unable to validate command: %w", err)
	}

	return l, nil
}

// WithCmd sets the command that will be run
func WithCmd(c string) commandOption {
	return func(l *command) error {
		l.Cmd = c
		return nil
	}
}

// WithArgs sets the arguments passed to the commands
func WithArgs(args []string) commandOption {
	return func(l *command) error {
		l.Args = args
		return nil
	}
}

// WithEnv sets the environment to run the command within
func WithEnv(env map[string]string) commandOption {
	return func(l *command) error {
		l.Env = env
		return nil
	}
}

// WithInheritedEnv automatically inherits the existing environment
func WithInheritedEnv() commandOption {
	return func(l *command) error {
		env := DefaultEnv()
		l.Env = env
		return nil
	}
}

// WithStdout sets the stdout
func WithStdout(fw io.Writer) commandOption {
	return func(l *command) error {
		l.Stdout = fw
		return nil
	}
}

// WithStdin sets the stderr
func WithStdin(fw io.Reader) commandOption {
	return func(l *command) error {
		l.Stdin = fw
		return nil
	}
}

// WithStderr sets the stderr
func WithStderr(fw io.Writer) commandOption {
	return func(l *command) error {
		l.Stderr = fw
		return nil
	}
}

// WithCwd sets cwd
func WithCwd(cwd string) commandOption {
	return func(l *command) error {
		l.Cwd = cwd
		return nil
	}
}
