package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/types/workflows/runnerhealthcheckbackfill"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const runnerHealthcheckBackfillNamespace = "general"

// @ID						AdminBackfillRunnerHealthcheckEmitters
// @Summary				backfill runner healthcheck emitters
// @Description			Starts a durable workflow that ensures every runner has a runner-healthcheck emitter on its runner-signals queue, creating (and starting the emitter workflow for) the ones that predate that behavior. Idempotent: runners that already have the emitter are skipped, and re-triggering reuses an in-progress backfill rather than starting a duplicate.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				202	{string}	ok
// @Router					/v1/general/backfill-runner-healthcheck-emitters [post]
func (s *service) AdminBackfillRunnerHealthcheckEmitters(ctx *gin.Context) {
	if err := s.startRunnerHealthcheckBackfill(ctx); err != nil {
		ctx.Error(fmt.Errorf("unable to start runner healthcheck backfill: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{"status": "ok"})
}

func (s *service) startRunnerHealthcheckBackfill(ctx context.Context) error {
	opts := tclient.StartWorkflowOptions{
		ID:                       runnerhealthcheckbackfill.WorkflowID,
		TaskQueue:                workflows.APITaskQueue,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		Memo: map[string]interface{}{
			"started-by": "ctl-api",
		},
	}

	if _, err := s.temporalClient.ExecuteWorkflowInNamespace(ctx, runnerHealthcheckBackfillNamespace, opts,
		runnerhealthcheckbackfill.WorkflowName, runnerhealthcheckbackfill.Request{}); err != nil {
		return fmt.Errorf("unable to execute runner healthcheck backfill workflow: %w", err)
	}

	return nil
}

type BackfillRunnerHealthcheckEmittersStatusResponse struct {
	Running        bool                                `json:"running"`
	Status         string                              `json:"status"`
	RunID          string                              `json:"run_id,omitempty"`
	StartedAt      *time.Time                          `json:"started_at,omitempty"`
	ClosedAt       *time.Time                          `json:"closed_at,omitempty"`
	ElapsedSeconds float64                             `json:"elapsed_seconds,omitempty"`
	Elapsed        string                              `json:"elapsed,omitempty"`
	Progress       *runnerhealthcheckbackfill.Progress `json:"progress,omitempty"`
}

// @ID						GetBackfillRunnerHealthcheckEmittersStatus
// @Summary				get runner healthcheck emitter backfill progress
// @Description			Reports whether the runner healthcheck emitter backfill workflow is still running and, when available, a live snapshot of how many runners have been processed and emitters created so far.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	BackfillRunnerHealthcheckEmittersStatusResponse
// @Router					/v1/general/backfill-runner-healthcheck-emitters [get]
func (s *service) GetBackfillRunnerHealthcheckEmittersStatus(ctx *gin.Context) {
	resp, err := s.getRunnerHealthcheckBackfillStatus(ctx)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("runner healthcheck backfill workflow has not been started")})
			return
		}
		ctx.Error(fmt.Errorf("unable to get runner healthcheck backfill status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getRunnerHealthcheckBackfillStatus(ctx context.Context) (*BackfillRunnerHealthcheckEmittersStatusResponse, error) {
	exec, err := s.temporalClient.DescribeWorkflowExecutionInNamespace(ctx, runnerHealthcheckBackfillNamespace, runnerhealthcheckbackfill.WorkflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe runner healthcheck backfill workflow: %w", err)
	}

	info := exec.GetWorkflowExecutionInfo()
	status := info.GetStatus()
	resp := &BackfillRunnerHealthcheckEmittersStatusResponse{
		Status:  status.String(),
		Running: status == enumsv1.WORKFLOW_EXECUTION_STATUS_RUNNING,
		RunID:   info.GetExecution().GetRunId(),
	}

	if startTS := info.GetStartTime(); startTS != nil {
		startedAt := startTS.AsTime()
		resp.StartedAt = &startedAt

		end := time.Now()
		if closeTS := info.GetCloseTime(); closeTS != nil {
			closedAt := closeTS.AsTime()
			resp.ClosedAt = &closedAt
			end = closedAt
		}

		elapsed := end.Sub(startedAt)
		resp.ElapsedSeconds = elapsed.Seconds()
		resp.Elapsed = elapsed.Round(time.Second).String()
	}

	encoded, err := s.temporalClient.QueryWorkflowInNamespace(ctx, runnerHealthcheckBackfillNamespace, runnerhealthcheckbackfill.WorkflowID, "", runnerhealthcheckbackfill.ProgressQueryType)
	if err != nil {
		// progress query is best-effort: a completed run on a closed workflow may not answer queries.
		return resp, nil
	}

	var progress runnerhealthcheckbackfill.Progress
	if err := encoded.Get(&progress); err != nil {
		return nil, fmt.Errorf("unable to decode runner healthcheck backfill progress: %w", err)
	}
	resp.Progress = &progress

	return resp, nil
}
