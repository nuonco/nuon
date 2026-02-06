package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// PolicyReportsQuery contains optional filters for listing policy reports.
type PolicyReportsQuery struct {
	OwnerType string
	OwnerID   string
	AppID     string
	InstallID string
	Status    string
	Offset    int
	Limit     int
}

// GetPolicyReports retrieves policy reports with optional filters.
func (c *client) GetPolicyReports(ctx context.Context, query *PolicyReportsQuery) ([]*models.AppPolicyReport, error) {
	reqURL := c.APIURL + "/v1/policy-reports"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if query != nil {
		if query.OwnerType != "" {
			q.Set("owner_type", query.OwnerType)
		}
		if query.OwnerID != "" {
			q.Set("owner_id", query.OwnerID)
		}
		if query.AppID != "" {
			q.Set("app_id", query.AppID)
		}
		if query.InstallID != "" {
			q.Set("install_id", query.InstallID)
		}
		if query.Status != "" {
			q.Set("status", query.Status)
		}
		if query.Offset > 0 {
			q.Set("offset", fmt.Sprintf("%d", query.Offset))
		}
		if query.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", query.Limit))
		}
	}
	req.URL.RawQuery = q.Encode()

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

	var reports []*models.AppPolicyReport
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return reports, nil
}

// GetPolicyReport retrieves a single policy report by ID.
func (c *client) GetPolicyReport(ctx context.Context, reportID string) (*models.AppPolicyReport, error) {
	reqURL := fmt.Sprintf("%s/v1/policy-reports/%s", c.APIURL, reportID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

	var report models.AppPolicyReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &report, nil
}

// ExportPolicyReport exports a policy report in the specified format.
// Valid formats: "json", "sarif", "pdf"
func (c *client) ExportPolicyReport(ctx context.Context, reportID, format string) ([]byte, string, error) {
	reqURL := fmt.Sprintf("%s/v1/policy-reports/%s/export", c.APIURL, reportID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if format != "" {
		q.Set("format", format)
	}
	req.URL.RawQuery = q.Encode()

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	return body, contentType, nil
}
