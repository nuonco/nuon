package start

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint"
	tclient "go.temporal.io/sdk/client"
)

type ProvisionInstanceRequest struct {
	OrgID        string             `json:"org_id" validate:"required"`
	AppID        string             `json:"app_id" validate:"required"`
	DeploymentID string             `json:"deployment_id" validate:"required"`
	InstallID    string             `json:"install_id" validate:"required"`
	Component    waypoint.Component `json:"component" validate:"required"`
}

func (p ProvisionInstanceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type ProvisionInstanceResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// provisioner exposes the methods needed to provision an instance
type provisioner interface {
	provisionInstance(context.Context, ProvisionInstanceRequest) (string, error)
}

var _ provisioner = (*instanceProvisioner)(nil)

type instanceProvisioner struct {
	TemporalHost      string
	TemporalNamespace string
}

func (a *Activities) ProvisionInstance(ctx context.Context, req ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
	resp := ProvisionInstanceResponse{}
	if err := req.validate(); err != nil {
		return resp, err
	}

	workflowID, err := a.provisionInstance(ctx, req)
	if err != nil {
		return resp, err
	}
	resp.WorkflowID = workflowID
	return resp, nil
}

func (i *instanceProvisioner) provisionInstance(ctx context.Context, req ProvisionInstanceRequest) (string, error) {
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
	req ProvisionInstanceRequest,
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
