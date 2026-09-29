package api

import (
	"context"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/runner/errcapture"
)

type resultCaptureClient struct {
	nuonrunner.Client
}

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
