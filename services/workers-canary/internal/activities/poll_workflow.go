package activities

import (
	"context"
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	temporalclient "github.com/powertoolsdev/mono/pkg/clients/temporal"
	"google.golang.org/protobuf/types/known/anypb"
)

//nolint:all
func (a *Activities) PollWorkflow(ctx context.Context, req *activitiesv1.PollWorkflowRequest) (*activitiesv1.PollWorkflowResponse, error) {
	tClient, err := temporalclient.New(a.v,
		temporalclient.WithNamespace(req.Namespace),
		temporalclient.WithAddr(a.TemporalHost))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	wkflow := tClient.GetWorkflow(ctx, req.WorkflowId, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow with id: %s: %w", req.WorkflowId, err)
	}

	// TODO(jm): I'm not a fan of this approach, but we can't pass an interface as an Any and then unmarshal it, so
	// there's not much of a better way TBH.
	var (
		resp      *anypb.Any
		wkflowErr error
	)
	switch [2]string{req.Namespace, req.WorkflowName} {
	case [2]string{"orgs", "Signup"}:
		var wkflowResp *orgsv1.SignupResponse
		if wkflowErr = wkflow.Get(ctx, &resp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
	case [2]string{"apps", "Provision"}:
		var wkflowResp *appsv1.ProvisionResponse
		if wkflowErr = wkflow.Get(ctx, &resp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
	case [2]string{"installs", "Provision"}:
		var wkflowResp *installsv1.ProvisionResponse
		if wkflowErr = wkflow.Get(ctx, &resp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
	case [2]string{"deployments", "Start"}:
		var wkflowResp *deploymentsv1.StartResponse
		if wkflowErr = wkflow.Get(ctx, &resp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", err)
		}
	}

	any, err := anypb.New(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to create any object from response: %w", err)
	}

	return &activitiesv1.PollWorkflowResponse{
		Step: &canaryv1.Step{
			// TODO(jm): fill this in
		},
		Response: any,
	}, nil
}
