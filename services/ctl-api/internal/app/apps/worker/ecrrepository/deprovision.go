package ecrrepository

import (
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
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
// @id-template {{.CallerID}}-deprovision-ecr-repo
func (w Wkflow) DeprovisionECRRepository(ctx workflow.Context, req *DeprovisionECRRepositoryRequest) (*DeprovisionECRRepositoryResponse, error) {
	l := log.With(workflow.GetLogger(ctx))

	l.Debug("destroying ecr repository", zap.String("noop", "true"))

	// TODO(jm): implement this
	return &DeprovisionECRRepositoryResponse{}, nil
}
