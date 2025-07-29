package kubernetes_manifest

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) writeErrorResult(ctx context.Context, err error) {
	resultReq := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:   false,
		ErrorCode: 0,
		ErrorMetadata: map[string]string{
			"handler":  h.Name(),
			"job_type": string(h.JobType()),
			"message":  err.Error(),
		},
	}

	if _, err := h.apiClient.CreateJobExecutionResult(
		ctx,
		h.state.jobID,
		h.state.jobExecutionID,
		resultReq,
	); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
}
