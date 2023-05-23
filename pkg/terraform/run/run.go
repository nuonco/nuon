package run

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// Run accepts a workspace, and executes the provided command in it, uploading outputs to the correct place, afterwards.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=run_mock.go -source=run.go -package=run
type Run interface {
	Apply(context.Context) error
	Plan(context.Context) error
	Destroy(context.Context) error
}

var _ Run = (*run)(nil)

type run struct {
	v *validator.Validate

	Workspace workspace.Workspace `validate:"required"`
}

type runOption func(*run) error

func New(v *validator.Validate, opts ...runOption) (*run, error) {
	r := &run{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("unable to set %d option: %w", idx, err)
		}
	}
	if err := r.v.Struct(r); err != nil {
		return nil, err
	}

	return r, nil
}

func WithWorkspace(w workspace.Workspace) runOption {
	return func(r *run) error {
		r.Workspace = w
		return nil
	}
}
