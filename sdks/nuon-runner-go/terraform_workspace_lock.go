package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
)

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
