package nuonrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) ReportCompositeErrors(ctx context.Context, jobID string, errors []models.CompositeError) error {
	reqBody := models.ReportCompositeErrorsRequest{
		Errors: errors,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("unable to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/runner-jobs/%s/errors", c.APIURL, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.appTransport.RoundTrip(httpReq)
	if err != nil {
		return fmt.Errorf("unable to report composite errors: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(respBody))
	}

	return nil
}
