package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
)

func (c *client) UpdateTerraformStateJSON(ctx context.Context, workspaceID string, jobID *string, reqBody any) (any, error) {
	resp, err := c.genClient.Operations.UpdateTerraformStateJSON(&operations.UpdateTerraformStateJSONParams{
		Body:        reqBody,
		Context:     ctx,
		WorkspaceID: workspaceID,
		JobID:       jobID,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
