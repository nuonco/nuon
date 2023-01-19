package instances

import (
	"context"

	provisionv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	tclient "go.temporal.io/sdk/client"
)

// provisioner exposes the methods needed to provision an instance
type provisioner interface {
	provisionInstance(context.Context, *provisionv1.ProvisionRequest) (string, error)
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

	_, err := a.provisionInstance(ctx, req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (i *instanceProvisioner) provisionInstance(ctx context.Context, req *provisionv1.ProvisionRequest) (string, error) {
	tc, err := tclient.Dial(tclient.Options{
		HostPort:  i.TemporalHost,
		Namespace: i.TemporalNamespace,
	})
	if err != nil {
		return "", err
	}

	return i.startWorkflow(ctx, tc, req)
}

func (i *instanceProvisioner) startWorkflow(
	ctx context.Context,
	client tclient.Client,
	req *provisionv1.ProvisionRequest,
) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "instance",
	}
	wkflow, err := client.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return "", err
	}
	return wkflow.GetID(), nil
}
