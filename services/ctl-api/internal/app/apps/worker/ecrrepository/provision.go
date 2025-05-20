package ecrrepository

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{.ParentID}}-provision-ecr-repo
func (w Wkflow) ProvisionECRRepository(ctx workflow.Context, req *ProvisionECRRepositoryRequest) (*ProvisionECRRepositoryResponse, error) {
	l := log.With(workflow.GetLogger(ctx))

	l.Debug("creating ecr repository")
	crReq := CreateRepositoryRequest{
		OrgID: req.OrgID,
		AppID: req.AppID,
	}
	ecrResp, err := AwaitCreateRepository(ctx, &crReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &ProvisionECRRepositoryResponse{
		RegistryID:     ecrResp.RegistryID,
		RepositoryARN:  ecrResp.RepositoryArn,
		RepositoryName: ecrResp.RepositoryName,
		RepositoryURI:  ecrResp.RepositoryURI,
	}, nil
}
