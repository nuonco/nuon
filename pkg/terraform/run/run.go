package run

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

type OutputSettings struct {
	Ignore         bool
	Credentials    *credentials.Config `validate:"required_unless=Ignore 1"`
	Bucket         string              `validate:"required_unless=Ignore 1"`
	JobPrefix      string              `validate:"required_unless=Ignore 1"`
	InstancePrefix string              `validate:"required_unless=Ignore 1"`
}

// Run accepts a workspace, and executes the provided command in it, uploading outputs to the correct place, afterwards.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=run_mock.go -source=run.go -package=run
type Run interface {
	Apply(context.Context) error
	Validate(context.Context) error
	Plan(context.Context) error
	Destroy(context.Context) error
}

var _ Run = (*run)(nil)

type run struct {
	v *validator.Validate

	Workspace      workspace.Workspace `validate:"required"`
	UI             terminal.UI         `validate:"required"`
	Log            hclog.Logger        `validate:"required"`
	OutputSettings *OutputSettings     `validate:"required"`
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

func WithOutputSettings(settings *OutputSettings) runOption {
	return func(r *run) error {
		r.OutputSettings = settings

		if err := r.v.Struct(settings); err != nil {
			return fmt.Errorf("unable to validate settings: %w", err)
		}

		return nil
	}
}

func WithWorkspace(w workspace.Workspace) runOption {
	return func(r *run) error {
		r.Workspace = w
		return nil
	}
}

func WithUI(ui terminal.UI) runOption {
	return func(r *run) error {
		r.UI = ui
		return nil
	}
}

func WithLogger(l hclog.Logger) runOption {
	return func(r *run) error {
		r.Log = l
		return nil
	}
}
