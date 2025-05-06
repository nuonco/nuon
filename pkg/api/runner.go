package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type RunnerServiceAccountTokenRequest struct {
	Duration   string `json:"duration"`
	Invalidate bool   `json:"invalidate"`
}

type RunnerServiceAccountTokenResponse struct {
	Token string `json:"token"`
}

func (c *client) GetRunnerServiceAccountToken(ctx context.Context, runnerID string, dur time.Duration, invalidate bool) (string, error) {
	endpoint := fmt.Sprintf("/v1/runners/%s/service-account-token", runnerID)
	byts, err := c.execPostRequest(ctx, endpoint, RunnerServiceAccountTokenRequest{
		Duration:   dur.String(),
		Invalidate: invalidate,
	})
	if err != nil {
		return "", fmt.Errorf("unable to execute post request: %w", err)
	}

	var resp RunnerServiceAccountTokenResponse
	if err := json.Unmarshal(byts, &resp); err != nil {
		return "", fmt.Errorf("unable to parse response: %w", err)
	}

	return resp.Token, nil
}

type RunnerGroup struct {
	Type    string `json:"type"`
	OwnerID string `json:"owner_id"`

	Settings *RunnerGroupSettings `json:"settings"`
}

type RunnerGroupSettings struct {
	LocalAWSIAMRoleARN string
}

type Runner struct {
	ID                string `json:"id"`
	OrgID             string `json:"org_id"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	RunnerGroupID string `json:"runner_group_id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
}

func (c *client) ListRunners(ctx context.Context, typ string) ([]Runner, error) {
	endpoint := "/v1/runners?type=" + typ
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute post request: %w", err)
	}

	var resp []Runner
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return resp, nil
}

func (c *client) GetRunner(ctx context.Context, id string) (*Runner, error) {
	endpoint := "/v1/runners/" + id
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var resp Runner
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &resp, nil
}

type RunnerServiceAccount struct {
	Email string `json:"email"`
}

func (c *client) GetRunnerServiceAccount(ctx context.Context, id string) (*RunnerServiceAccount, error) {
	endpoint := "/v1/runners/" + id + "/service-account"
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var resp RunnerServiceAccount
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to parse response")
	}

	return &resp, nil
}

func (c *client) GetRunnerGroup(ctx context.Context, id string) (*RunnerGroup, error) {
	endpoint := "/v1/runner-groups/" + id
	byts, err := c.execGetRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to execute get request: %w", err)
	}

	var resp RunnerGroup
	if err := json.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &resp, nil
}

func (c *client) RestartRunner(ctx context.Context, runnerID string) error {
	endpoint := fmt.Sprintf("/v1/runners/%s/restart", runnerID)
	byts, err := c.execPostRequest(ctx, endpoint, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("unable to execute post request: %w", err)
	}

	var response bool
	if err := json.Unmarshal(byts, &response); err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}
	if !response {
		return fmt.Errorf("unable to restart runner: %v", response)
	}

	return nil
}
