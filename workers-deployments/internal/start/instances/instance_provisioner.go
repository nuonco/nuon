package instances

import (
	"context"
	"fmt"

	provisionv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	tclient "go.temporal.io/sdk/client"
)

// provisioner exposes the methods needed to provision an instance
type provisioner interface {
	provisionInstance(context.Context, *provisionv1.ProvisionRequest) error
}

var _ provisioner = (*instanceProvisioner)(nil)

type instanceProvisioner struct {
	TemporalHost      string
	TemporalNamespace string
}

func (a *Activities) ProvisionInstance(ctx context.Context, req *provisionv1.ProvisionRequest) (*provisionv1.ProvisionResponse, error) {
	resp := &provisionv1.ProvisionResponse{}
	if err := req.Validate(); err != nil {
		return resp, err
	}

	err := a.provisionInstance(ctx, req)
	if err != nil {
		return resp, fmt.Errorf("failed to provision instance: %w", err)
	}
	return resp, nil
}

func (i *instanceProvisioner) provisionInstance(ctx context.Context, req *provisionv1.ProvisionRequest) error {
	tc, err := tclient.Dial(tclient.Options{
		HostPort:  i.TemporalHost,
		Namespace: i.TemporalNamespace,
	})
	if err != nil {
		return err
	}

	return i.startWorkflow(ctx, tc, req)
}

func (i *instanceProvisioner) startWorkflow(
	ctx context.Context,
	client tclient.Client,
	req *provisionv1.ProvisionRequest,
) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "instance",
		Memo: map[string]interface{}{
			"org-id":        req.OrgId,
			"app-id":        req.AppId,
			"deployment-id": req.DeploymentId,
			"install-id":    req.InstallId,
		},
	}
	wkflow, err := client.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to submit workflow: %w", err)
	}

	resp := &sharedv1.Response{}

	// NOTE(jm): we wait for the workflow here, to ensure that the deployment workflow is _actually_ done when it
	// finishes
	if err := wkflow.Get(ctx, resp); err != nil {
		return fmt.Errorf("workflow failed: %w", err)
	}

	return nil
}
