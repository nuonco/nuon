package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) TaintRunner(ctx context.Context, runnerID string) (*models.AppRunner, error) {
	return c.postRunnerAction(ctx, runnerID, "taint")
}

func (c *client) UntaintRunner(ctx context.Context, runnerID string) (*models.AppRunner, error) {
	return c.postRunnerAction(ctx, runnerID, "untaint")
}

func (c *client) postRunnerAction(ctx context.Context, runnerID, action string) (*models.AppRunner, error) {
	reqURL := fmt.Sprintf("%s/v1/runners/%s/%s", c.APIURL, runnerID, action)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var runner models.AppRunner
	if err := json.NewDecoder(resp.Body).Decode(&runner); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &runner, nil
}
