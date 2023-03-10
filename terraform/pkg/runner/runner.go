package runner

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

type runner struct {
	Plan *planv1.TerraformPlan `validate:"required,dive"`

	// internal state
	validator        *validator.Validate
	cleanupFns       []func() error
	workspaceSetuper workspaceSetuper
}

type runnerOption func(*runner) error

// New instantiates a new runner
func New(v *validator.Validate, opts ...runnerOption) (*runner, error) {
	r := &runner{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating runner: validator is nil")
	}
	r.validator = v

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := r.validator.Struct(r); err != nil {
		return nil, err
	}

	r.workspaceSetuper = r

	return r, nil
}

// WithPlan specifies the terraform plan to execute
func WithPlan(p *planv1.TerraformPlan) runnerOption {
	return func(r *runner) error {
		r.Plan = p
		return nil
	}
}

// cleanup runs the cleanup functions for the runner and returns the consolidated errors
// safe to run even if there are no cleanupFns
func (r *runner) cleanup() error {
	var err error
	for _, fn := range r.cleanupFns {
		e := fn()
		if e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}
