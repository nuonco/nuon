package command

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	"github.com/fatih/color"

	"github.com/go-playground/validator/v10"
)

type command struct {
	v *validator.Validate

	LinePrefixFn   func() string
	LinePrefix     string
	LineColor      *color.Color
	FileOutputPath string

	Cmd  string `validate:"required"`
	Args []string
	Env  map[string]string `validate:"required"`

	Cwd    string
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer `validate:"required"`

	UseProcessGroup bool
}

type commandOption func(*command) error

func New(v *validator.Validate, opts ...commandOption) (*command, error) {
	l := &command{
		v:      v,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Env:    make(map[string]string, 0),
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

func WithCmd(c string) commandOption {
	return func(l *command) error {
		l.Cmd = c
		return nil
	}
}

func WithArgs(args []string) commandOption {
	return func(l *command) error {
		l.Args = args
		return nil
	}
}

func WithEnv(env map[string]string) commandOption {
	return func(l *command) error {
		for k, v := range env {
			l.Env[k] = v
		}

		return nil
	}
}

func WithInheritedEnv() commandOption {
	return func(l *command) error {
		env := DefaultEnv()
		l.Env = env
		return nil
	}
}

func WithStdout(fw io.Writer) commandOption {
	return func(l *command) error {
		l.Stdout = fw
		return nil
	}
}

func WithStdin(fw io.Reader) commandOption {
	return func(l *command) error {
		l.Stdin = fw
		return nil
	}
}

func WithStderr(fw io.Writer) commandOption {
	return func(l *command) error {
		l.Stderr = fw
		return nil
	}
}

func WithCwd(cwd string) commandOption {
	return func(l *command) error {
		l.Cwd = cwd
		return nil
	}
}

func WithProcessGroup() commandOption {
	return func(l *command) error {
		l.UseProcessGroup = true
		return nil
	}
}

func WithLinePrefix(prefix string) commandOption {
	return func(l *command) error {
		l.LinePrefix = prefix
		return nil
	}
}

func WithLinePrefixFn(fn func() string) commandOption {
	return func(l *command) error {
		l.LinePrefixFn = fn
		return nil
	}
}

func WithLineColor(color *color.Color) commandOption {
	return func(l *command) error {
		l.LineColor = color
		return nil
	}
}

func WithFileOutput(fp string) commandOption {
	return func(l *command) error {
		l.FileOutputPath = fp
		return nil
	}
}

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func WithTTYAwareStdin() commandOption {
	return func(l *command) error {
		if IsTTY() {
			l.Stdin = os.Stdin
		} else {
			l.Stdin = io.NopCloser(nil)
		}
		return nil
	}
}

func WithTTYAwareEnv(env map[string]string) commandOption {
	return func(l *command) error {
		l.Env = DefaultEnv()

		for k, v := range env {
			l.Env[k] = v
		}

		return nil
	}
}
