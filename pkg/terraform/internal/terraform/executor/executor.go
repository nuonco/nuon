package executor

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/terraform-exec/tfexec"
	"go.uber.org/zap"
)

type tfExecutor struct {
	WorkingDir        string `validate:"required,dir"`
	ExecPath          string `validate:"required,file"`
	BackendConfigFile string `validate:"required"`
	VarFile           string `validate:"required"`

	// internal state
	validator *validator.Validate
	initer    initer
	planner   planner
	applier   applier
	destroyer destroyer
	outputter outputter
}

type tfExecutorOption func(*tfExecutor) error

// New instantiates a new terraform executor
func New(v *validator.Validate, opts ...tfExecutorOption) (*tfExecutor, error) {
	e := &tfExecutor{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating configurator: validator is nil")
	}
	e.validator = v

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	if err := e.validator.Struct(e); err != nil {
		return nil, err
	}

	tfClient, err := tfexec.NewTerraform(e.WorkingDir, e.ExecPath)
	if err != nil {
		// NOTE(jdt): we validate the inputs before this so this should never error
		return nil, err
	}

	// TODO(jdt): plumb logger through instead of using zap.L
	tfClient.SetLogger(&printfer{l: zap.L()})
	tfClient.SetStderr(os.Stderr)
	tfClient.SetStdout(os.Stdout)

	e.initer = tfClient
	e.planner = tfClient
	e.applier = tfClient
	e.destroyer = tfClient
	e.outputter = tfClient

	return e, nil
}

// WithWorkingDir sets the working directory where the module and config lives
func WithWorkingDir(d string) tfExecutorOption {
	return func(te *tfExecutor) error {
		te.WorkingDir = d
		return nil
	}
}

// WithTerraformExecPath sets the path to the installed terraform binary
func WithTerraformExecPath(d string) tfExecutorOption {
	return func(te *tfExecutor) error {
		te.ExecPath = d
		return nil
	}
}

// WithBackendConfigFile specifies the backend config file to use
func WithBackendConfigFile(f string) tfExecutorOption {
	return func(te *tfExecutor) error {
		te.BackendConfigFile = f
		return nil
	}
}

// WithVarFile specifies the var file.
// May be empty but must be present
func WithVarFile(f string) tfExecutorOption {
	return func(te *tfExecutor) error {
		te.VarFile = f
		return nil
	}
}

type printfer struct {
	l *zap.Logger
}

func (p *printfer) Printf(format string, v ...interface{}) {
	p.l.Sugar().Infof(format, v...)
}
