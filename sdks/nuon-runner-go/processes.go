package nuonrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) CreateProcess(ctx context.Context, req *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.CreateRunnerProcess(&operations.CreateRunnerProcessParams{
		RunnerID: c.RunnerID,
		Req:      req,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetProcess(ctx context.Context, processID string) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.GetRunnerProcess(&operations.GetRunnerProcessParams{
		RunnerID:  c.RunnerID,
		ProcessID: processID,
		Context:   ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateProcess(ctx context.Context, processID string, req *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.UpdateRunnerProcess(&operations.UpdateRunnerProcessParams{
		RunnerID:  c.RunnerID,
		ProcessID: processID,
		Req:       req,
		Context:   ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CompleteShutdown(ctx context.Context, processID, shutdownID string) (*models.AppRunnerProcessShutdown, error) {
	url := fmt.Sprintf("%s/v1/runners/%s/processes/%s/shutdowns/%s/complete", c.APIURL, c.RunnerID, processID, shutdownID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	httpResp, err := c.appTransport.RoundTrip(httpReq)
	if err != nil {
		return nil, fmt.Errorf("unable to complete shutdown: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(body))
	}

	var shutdown models.AppRunnerProcessShutdown
	if err := json.Unmarshal(body, &shutdown); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return &shutdown, nil
}
