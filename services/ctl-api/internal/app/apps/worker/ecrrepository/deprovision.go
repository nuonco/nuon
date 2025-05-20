package ecrrepository

import (
	"go.temporal.io/sdk/workflow"
)

type DeprovisionECRRepositoryRequest struct {
	OrgID string
	AppID string

	WorkflowID string `validate:"required"`
}

type DeprovisionECRRepositoryResponse struct{}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{.ParentID}}-deprovision-ecr-repo
func (w Wkflow) DeprovisionECRRepository(ctx workflow.Context, req *DeprovisionECRRepositoryRequest) (*DeprovisionECRRepositoryResponse, error) {
	// TODO(jm): implement this
	return nil, nil
}
