package nuonrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type jobTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *client) ClaimJobToken(ctx context.Context, jobID string) (string, error) {
	reqURL := fmt.Sprintf("%s/v1/runners/%s/jobs/%s/token", c.APIURL, c.RunnerID, jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create job token request: %w", err)
	}
	client := &http.Client{
		Transport: c.appTransport,
		Timeout:   c.RequestTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to claim job token: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("unable to read job token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claim job token: status %d: %s", resp.StatusCode, string(body))
	}
	var parsed jobTokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unable to decode job token: %w", err)
	}
	if parsed.Token == "" {
		return "", fmt.Errorf("claim job token returned an empty token")
	}
	return parsed.Token, nil
}
