package nuon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) BuildAppConfig(ctx context.Context, appID, configID string, requestID ...string) (*models.AppWorkflow, error) {
	reqURL := fmt.Sprintf("%s/v1/apps/%s/configs/%s/build", c.APIURL, appID, configID)

	var body io.Reader
	if id := optionalRequestID(requestID); id != "" {
		payload, err := json.Marshal(&models.ServiceBuildAppConfigRequest{RequestID: id})
		if err != nil {
			return nil, fmt.Errorf("failed to encode request: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var workflow models.AppWorkflow
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workflow, nil
}
