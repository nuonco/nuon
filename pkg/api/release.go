package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type Release struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *client) ListReleases(ctx context.Context) ([]Release, error) {
	endpoint := "/v1/releases"
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var response []Release
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return response, nil
}

func (c *client) RestartRelease(ctx context.Context, releaseID string) error {
	endpoint := fmt.Sprintf("/v1/releases/%s/admin-restart", releaseID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restartrelease: %v", response)
	}

	return nil
}
