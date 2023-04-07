package temporal

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_repo.go -source=repo.go -package=temporal
type Repo interface {
	TriggerCanaryProvision(context.Context, *canaryv1.ProvisionRequest) error

	TriggerDeploymentStart(context.Context, *deploymentsv1.StartRequest) error
	ExecDeploymentStart(context.Context, *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error)

	TriggerInstallProvision(context.Context, *installsv1.ProvisionRequest) error
	TriggerInstallDeprovision(context.Context, *installsv1.DeprovisionRequest) error

	TriggerOrgSignup(context.Context, *orgsv1.SignupRequest) error
	ExecOrgSignup(context.Context, *orgsv1.SignupRequest) (*orgsv1.SignupResponse, error)
}

type repo struct {
	v *validator.Validate

	Client temporal.Client `validate:"required"`
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

func WithClient(client temporal.Client) repoOption {
	return func(r *repo) error {
		r.Client = client
		return nil
	}
}
