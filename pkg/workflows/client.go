package workflows

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=workflows
type Client interface {
	TriggerCanaryProvision(context.Context, *canaryv1.ProvisionRequest) error
	ScheduleCanaryProvision(context.Context, string, string, *canaryv1.ProvisionRequest) error
	UnscheduleCanaryProvision(context.Context, string) error
	TriggerCanaryDeprovision(context.Context, *canaryv1.DeprovisionRequest) error

	TriggerDeploymentStart(context.Context, *deploymentsv1.StartRequest) error
	ExecDeploymentStart(context.Context, *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error)

	TriggerInstallProvision(context.Context, *installsv1.ProvisionRequest) error
	TriggerInstallDeprovision(context.Context, *installsv1.DeprovisionRequest) error

	TriggerOrgSignup(context.Context, *orgsv1.SignupRequest) error
	ExecOrgSignup(context.Context, *orgsv1.SignupRequest) (*orgsv1.SignupResponse, error)
	TriggerOrgTeardown(context.Context, *orgsv1.TeardownRequest) error

	ExecCreatePlan(ctx context.Context, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error)
}

type workflowsClient struct {
	v *validator.Validate

	TemporalClient temporal.Client `validate:"required"`
}

var _ Client = (*workflowsClient)(nil)

// New returns a default repo with the default orgcontext getter
func NewClient(v *validator.Validate, opts ...workflowsClientOption) (*workflowsClient, error) {
	r := &workflowsClient{
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

type workflowsClientOption func(*workflowsClient) error

func WithClient(tclient temporal.Client) workflowsClientOption {
	return func(r *workflowsClient) error {
		r.TemporalClient = tclient
		return nil
	}
}
