package activities

import (
	"context"
	"fmt"
	"time"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	buildsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PollActivityTimeout = time.Minute * 1
	MaxActivityRetries  = 50
	DefaultRegion       = "us-west-2"
)

func (a *Activities) PollWorkflow(ctx context.Context, req *activitiesv1.PollWorkflowRequest) (*activitiesv1.PollWorkflowResponse, error) {
	tClient, err := temporalclient.New(a.v,
		temporalclient.WithNamespace(req.Namespace),
		temporalclient.WithAddr(a.TemporalHost))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	wkflow, err := tClient.GetWorkflowInNamespace(ctx, req.Namespace, req.WorkflowId, "")
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
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"orgs", "Teardown"}:
		var wkflowResp *orgsv1.TeardownResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"apps", "Provision"}:
		var wkflowResp *appsv1.ProvisionResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"installs", "Provision"}:
		var wkflowResp *installsv1.ProvisionResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"installs", "Deprovision"}:
		var wkflowResp *installsv1.DeprovisionResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"deployments", "Start"}:
		var wkflowResp *deploymentsv1.StartResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	case [2]string{"builds", "Build"}:
		var wkflowResp *buildsv1.BuildResponse
		if wkflowErr = wkflow.Get(ctx, &wkflowResp); wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
		resp, wkflowErr = anypb.New(wkflowResp)
		if wkflowErr != nil {
			return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
		}
	}

	return &activitiesv1.PollWorkflowResponse{
		Response: resp,
	}, nil
}
