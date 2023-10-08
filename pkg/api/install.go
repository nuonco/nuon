package api

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *client) ForgetInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-forget", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to forget install: %v", response)
	}

	return nil
}

func (c *client) ReprovisionInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-reprovision", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to reprovision install: %v", response)
	}

	return nil
}

func (c *client) RestartInstall(ctx context.Context, installID string) error {
	endpoint := fmt.Sprintf("/v1/installs/%s/admin-restart", installID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restart install: %v", response)
	}

	return nil
}
