package api

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *client) GetFeatures(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("/v1/orgs/admin-features")
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var response []string
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return response, nil
}

func (c *client) UpdateOrgFeatures(ctx context.Context, orgID string, features map[string]bool) error {
	body := map[string]interface{}{
		"features": features,
	}

	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-features", orgID)
	_, err := c.execPatchRequest(ctx, endpoint, body)
	if err != nil {
		return fmt.Errorf("unable to execute request: %w", err)
	}

	return nil
}
