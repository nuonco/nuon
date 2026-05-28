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

// GetTerraformWorkspaceLatestState fetches the latest state with contents by first
// listing states (which omits Contents) then fetching by ID (which includes Contents).
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

// GetTerraformWorkspaceLatestStateJSON fetches the latest state-json (terraform show -json format)
// by listing state-json records then fetching by ID. The by-ID endpoint returns the raw
// terraform show -json output directly, so we return it as json.RawMessage rather than re-marshaling
// the generated client's typed payload — callers consume it as raw JSON.
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
