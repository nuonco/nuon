package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type RunnerServiceAccountTokenRequest struct {
	Duration string `json:"duration"`
}

type RunnerServiceAccountTokenResponse struct {
	Token string `json:"token"`
}

func (c *client) GetRunnerServiceAccountToken(ctx context.Context, runnerID string, dur time.Duration) (string, error) {
	endpoint := fmt.Sprintf("/v1/runners/%s/service-account-token", runnerID)
	byts, err := c.execPostRequest(ctx, endpoint, RunnerServiceAccountTokenRequest{
		Duration: dur.String(),
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

type Runner struct {
	ID                string `json:"id"`
	OrgID             string `json:"org_id"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	RunnerGroupID string `json:"runner_group_id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
}

func (c *client) ListRunners(ctx context.Context) ([]Runner, error) {
	endpoint := "/v1/runners"
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
