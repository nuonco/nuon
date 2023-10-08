package api

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *client) DeleteOrg(ctx context.Context, orgID string) error {
	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-delete", orgID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to delete org: %v", response)
	}

	return nil
}

func (c *client) ReprovisionOrg(ctx context.Context, orgID string) error {
	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-reprovision", orgID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to reprovision org: %v", response)
	}

	return nil
}

func (c *client) RestartOrg(ctx context.Context, orgID string) error {
	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-restart", orgID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restart org: %v", response)
	}

	return nil
}
