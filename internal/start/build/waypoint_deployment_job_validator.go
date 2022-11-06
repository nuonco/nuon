package build

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type ValidateWaypointDeploymentJobRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	JobID string `json:"job_id" validate:"required"`
}

type ValidateWaypointDeploymentJobResponse struct{}

func (p ValidateWaypointDeploymentJobRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

func (a *Activities) ValidateWaypointDeploymentJob(
	ctx context.Context,
	req ValidateWaypointDeploymentJobRequest,
) (ValidateWaypointDeploymentJobResponse, error) {
	var resp ValidateWaypointDeploymentJobResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate waypoint deploy job: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.validateWaypointDeploymentJob(ctx, client, req.JobID); err != nil {
		return resp, fmt.Errorf("unable to validate job: %w", err)
	}

	return resp, nil
}

var _ waypointDeploymentJobValidator = (*waypointDeploymentJobValidatorImpl)(nil)

type waypointDeploymentJobValidatorImpl struct{}

func (waypointDeploymentJobValidatorImpl) validateWaypointDeploymentJob(ctx context.Context, client waypointClientJobValidator, jobID string) error {
	resp, err := client.ValidateJob(ctx, &gen.ValidateJobRequest{
		Job: &gen.Job{
			Id: jobID,
		},
	})
	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

type waypointDeploymentJobValidator interface {
	validateWaypointDeploymentJob(context.Context, waypointClientJobValidator, string) error
}

type waypointClientJobValidator interface {
	ValidateJob(ctx context.Context, in *gen.ValidateJobRequest, opts ...grpc.CallOption) (*gen.ValidateJobResponse, error)
}
