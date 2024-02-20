package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Org struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *client) ListOrgs(ctx context.Context) ([]Org, error) {
	endpoint := "/v1/orgs"
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var response []Org
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return response, nil
}

func (c *client) GetOrg(ctx context.Context, nameOrID string) (*Org, error) {
	endpoint := "/v1/orgs/admin-get?name=" + nameOrID
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var response Org
	if err := json.Unmarshal(byts, &response); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &response, nil
}

func (c *client) DeleteOrg(ctx context.Context, orgID string) error {
	origTimeout := c.Timeout
	defer func() {
		c.Timeout = origTimeout
	}()
	c.Timeout = time.Second * 10

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

func (c *client) DeprovisionOrg(ctx context.Context, orgID string) error {
	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-deprovision", orgID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to deprovision org: %v", response)
	}

	return nil
}

func (c *client) ForgetOrgInstalls(ctx context.Context, orgID string) error {
	endpoint := fmt.Sprintf("/v1/orgs/%s/admin-forget-installs", orgID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to forget org installs: %v", response)
	}

	return nil
}
