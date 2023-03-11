package executors

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal"
)

const (
	assumeRoleSessionName string = "nuonctl-executors"
)

type Repo interface {
	GetPlan(context.Context, *planv1.PlanRef) (*planv1.Plan, error)
	GetDeploymentsPlan(context.Context, string) (*planv1.Plan, error)
}

type repo struct {
	v *validator.Validate

	IAMRoleARN        string `validate:"required"`
	DeploymentsBucket string `validate:"required"`
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

func WithConfig(cfg *internal.Config) repoOption {
	return func(r *repo) error {
		r.IAMRoleARN = cfg.SupportIAMRoleArn
		r.DeploymentsBucket = cfg.DeploymentsBucket
		return nil
	}
}
