package temporal

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	orgsv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1"
	tclient "go.temporal.io/sdk/client"
)

type Repo interface {
	TriggerDeploymentStart(context.Context, *deploymentsv1.StartRequest) error
	ExecDeploymentStart(context.Context, *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error)

	TriggerOrgSignup(context.Context, *orgsv1.SignupRequest) error
	ExecOrgSignup(context.Context, *orgsv1.SignupRequest) (*orgsv1.SignupResponse, error)
}

type repo struct {
	v *validator.Validate

	Client tclient.Client `validate:"required"`
}

var _ Repo = (*repo)(nil)

// New returns a default repo with the default orgcontext getter
func New(v *validator.Validate, opts ...repoOption) (*repo, error) {
	r := &repo{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate repo: %w", err)
	}

	return r, nil
}

type repoOption func(*repo) error

func WithClient(client tclient.Client) repoOption {
	return func(r *repo) error {
		r.Client = client
		return nil
	}
}
