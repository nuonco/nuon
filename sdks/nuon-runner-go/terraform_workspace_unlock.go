package nuonrunner

import (
	"context"
	"encoding/json"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
)

func (c *client) UnlockTerraformWorkspace(ctx context.Context, workspaceID string) error {
	_, err := c.genClient.Operations.UnlockTerraformWorkspace(&operations.UnlockTerraformWorkspaceParams{
		Body:        json.RawMessage(`{}`),
		Context:     ctx,
		WorkspaceID: workspaceID,
	}, c.getAuthInfo())
	if err != nil {
		return err
	}

	return nil
}
