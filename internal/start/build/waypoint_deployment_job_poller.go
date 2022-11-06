package build

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint/job"
	"go.temporal.io/sdk/activity"
)

type PollWaypointDeploymentJobRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	BucketName   string `json:"bucket_name" validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`

	JobID string `json:"job_id" validate:"required"`
}

type PollWaypointDeploymentJobResponse struct{}

func (p PollWaypointDeploymentJobRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

func (a *Activities) PollWaypointDeploymentJob(
	ctx context.Context,
	req PollWaypointDeploymentJobRequest,
) (PollWaypointDeploymentJobResponse, error) {
	var resp PollWaypointDeploymentJobResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate waypoint deploy job: %w", err)
	}

	l := activity.GetLogger(ctx)

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	writer := newLogEventWriter(l)
	if err := job.Poll(ctx, client, req.JobID, writer); err != nil {
		return resp, fmt.Errorf("unable to finish waypoint deployment job: %w", err)
	}

	return resp, nil
}
