package terraform

import (
	"context"

	composite_errors "github.com/nuonco/nuon/bins/runner/internal/pkg/composite_errors"
	"github.com/nuonco/nuon/pkg/terraform/run"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) writeErrorResult(ctx context.Context, step string, err error) {
	resultReq := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:   false,
		ErrorCode: 0,
		ErrorMetadata: map[string]string{
			"step":     step,
			"handler":  h.Name(),
			"job_type": string(h.JobType()),
			"message":  err.Error(),
		},
	}

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
}

func (h *handler) reportCompositeErrors(ctx context.Context, tfRun run.Run, ownerType string, runErr error) {
	outputBytes := tfRun.LastOutputBytes()

	var errs []composite_errors.CompositeError
	if len(outputBytes) > 0 {
		errs = composite_errors.ParseTerraformJSON(outputBytes, ownerType)
	}

	// Fallback: wrap the Go error if no structured diagnostics were found
	if len(errs) == 0 {
		errs = composite_errors.FromGoError(runErr, ownerType)
	}

	modelErrs := composite_errors.ToModels(errs)
	if err := h.apiClient.ReportCompositeErrors(ctx, h.state.jobID, modelErrs); err != nil {
		h.errRecorder.Record("report composite errors", err)
	}
}
