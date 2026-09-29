package nuon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetTerraformWorkspaceStatesJSON(ctx context.Context, workspaceID string) ([]*models.AppTerraformWorkspaceStateJSON, error) {
	limit := int64(1)
	resp, err := c.genClient.Operations.GetTerraformWorkspaceStatesJSONV2(
		&operations.GetTerraformWorkspaceStatesJSONV2Params{
			WorkspaceID: workspaceID,
			Limit:       &limit,
			Context:     ctx,
		},
		c.getOrgIDAuthInfo(),
	)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

func (c *client) GetTerraformWorkspaceStates(ctx context.Context, workspaceID string) ([]*models.AppTerraformWorkspaceState, error) {
	limit := int64(1)
	resp, err := c.genClient.Operations.GetTerraformStatesV2(
		&operations.GetTerraformStatesV2Params{
			WorkspaceID: workspaceID,
			Limit:       &limit,
			Context:     ctx,
		},
		c.getOrgIDAuthInfo(),
	)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

func (c *client) GetTerraformWorkspaceLatestState(ctx context.Context, workspaceID string) (*models.AppTerraformWorkspaceState, error) {
	states, err := c.GetTerraformWorkspaceStates(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing states: %w", err)
	}
	if len(states) == 0 {
		return nil, nil
	}

	resp, err := c.genClient.Operations.GetTerraformWorkspaceStateByIDV2(
		&operations.GetTerraformWorkspaceStateByIDV2Params{
			WorkspaceID: workspaceID,
			StateID:     states[0].ID,
			Context:     ctx,
		},
		c.getOrgIDAuthInfo(),
	)
	if err != nil {
		return nil, fmt.Errorf("fetching state by ID: %w", err)
	}
	return resp.GetPayload(), nil
}

func (c *client) GetTerraformWorkspaceLatestStateJSON(ctx context.Context, workspaceID string) (json.RawMessage, error) {
	states, err := c.GetTerraformWorkspaceStatesJSON(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing state-json: %w", err)
	}
	if len(states) == 0 {
		return nil, nil
	}

	resp, err := c.genClient.Operations.GetTerraformWorkspaceStatesJSONByIDV2(
		&operations.GetTerraformWorkspaceStatesJSONByIDV2Params{
			WorkspaceID: workspaceID,
			StateID:     states[0].ID,
			Context:     ctx,
		},
		c.getOrgIDAuthInfo(),
	)
	if err != nil {
		return nil, fmt.Errorf("fetching state-json by ID: %w", err)
	}
	raw, err := json.Marshal(resp.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("marshaling state-json payload: %w", err)
	}
	return raw, nil
}
