package api

import (
	"context"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/errcapture"
)

// resultCaptureClient decorates a nuonrunner.Client to attach the execution's
// captured error output to failed job results. Every result write in the runner
// goes through CreateJobExecutionResult, so this is the single place that needs
// to know about error capture — handlers stay unaware of it.
type resultCaptureClient struct {
	nuonrunner.Client
}

// CreateJobExecutionResult attaches the captured error output (from the job
// logger's error-capture core) to a failed result under errcapture.MetadataKey,
// unless the handler already set it. Successful results are passed through
// untouched.
func (c *resultCaptureClient) CreateJobExecutionResult(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	if req != nil && !req.Success {
		if out := errcapture.Output(ctx); out != "" {
			if req.ErrorMetadata == nil {
				req.ErrorMetadata = map[string]string{}
			}
			if req.ErrorMetadata[errcapture.MetadataKey] == "" {
				req.ErrorMetadata[errcapture.MetadataKey] = out
			}
		}
	}
	return c.Client.CreateJobExecutionResult(ctx, jobID, jobExecutionID, req)
}
