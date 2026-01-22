package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetTerraformWorkspaceLock(ctx context.Context, workspaceID string) (*models.AppTerraformWorkspaceLock, error) {
	res, err := c.genClient.Operations.GetTerraformWorkspaceLock(&operations.GetTerraformWorkspaceLockParams{
		Context:     ctx,
		WorkspaceID: workspaceID,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return res.Payload, nil
}

func (c *client) LockTerraformWorkspace(ctx context.Context, workspaceID string, jobID *string, reqBody any) error {
	_, err := c.genClient.Operations.LockTerraformWorkspace(&operations.LockTerraformWorkspaceParams{
		Body:        reqBody,
		Context:     ctx,
		WorkspaceID: workspaceID,
		JobID:       jobID,
	}, c.getAuthInfo())
	if err != nil {
		return err
	}

	return nil
}
