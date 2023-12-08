package client

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	temporal "github.com/powertoolsdev/mono/pkg/temporal/client"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
)

const defaultAgent = "unknown"

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=client
type Client interface {
	TriggerInstallProvision(context.Context, *installsv1.ProvisionRequest) (string, error)
	TriggerInstallDeprovision(context.Context, *installsv1.DeprovisionRequest) (string, error)

	TriggerOrgSignup(context.Context, *orgsv1.ProvisionRequest) (string, error)
	ExecOrgSignup(context.Context, *orgsv1.ProvisionRequest) (*orgsv1.ProvisionResponse, error)
	TriggerOrgTeardown(context.Context, *orgsv1.DeprovisionRequest) (string, error)

	ExecCreatePlan(ctx context.Context, namespace string, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error)
	ExecExecutePlan(ctx context.Context, namespace string, req *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error)

	TriggerAppProvision(context.Context, *appsv1.ProvisionRequest) (string, error)
}

type workflowsClient struct {
	v *validator.Validate

	TemporalClient temporal.Client `validate:"required"`
	Agent          string          `validate:"required"`
}

var _ Client = (*workflowsClient)(nil)

// New returns a default repo with the default orgcontext getter
func NewClient(v *validator.Validate, opts ...workflowsClientOption) (*workflowsClient, error) {
	r := &workflowsClient{
		v:     v,
		Agent: defaultAgent,
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

func WithAgent(agent string) workflowsClientOption {
	return func(r *workflowsClient) error {
		r.Agent = agent
		return nil
	}
}
