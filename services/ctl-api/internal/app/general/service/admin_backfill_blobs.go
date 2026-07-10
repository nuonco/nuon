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

	"github.com/nuonco/nuon/pkg/types/workflows/blobbackfill"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const blobBackfillNamespace = "general"

// AdminBackfillBlobsRequest optionally narrows the backfill to a subset of
// dual-write tables. When empty, every supported table is processed.
type AdminBackfillBlobsRequest struct {
	Tables []string `json:"tables"`
}

// @ID						AdminBackfillBlobs
// @Summary				backfill dual-write blobs to S3
// @Description			Starts an orchestrator workflow that enumerates the day-buckets of existing dual-write rows and mirrors each day's rows into S3 blob storage, throttled to the configured per-second S3 write rate. Re-running drains any remaining backlog; an already-running backfill is reused rather than duplicated.
// @Description
// @Description			Supported tables: install_states, install_workflow_step_approvals, runner_job_plans, runner_job_execution_outputs, terraform_workspace_states, terraform_workspace_state_jsons. Leave `tables` empty (or omit the body) to process all of them.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Param					request	body	AdminBackfillBlobsRequest	false	"optional table subset; empty processes all supported tables"
// @Success				201	{string}	ok
// @Router					/v1/general/backfill-blobs [post]
func (s *service) AdminBackfillBlobs(ctx *gin.Context) {
	var req AdminBackfillBlobsRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(fmt.Errorf("unable to parse request: %w", err))
			return
		}
	}

	if err := s.startBlobBackfill(ctx, req.Tables); err != nil {
		ctx.Error(fmt.Errorf("unable to start blob backfill: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) startBlobBackfill(ctx context.Context, tables []string) error {
	opts := tclient.StartWorkflowOptions{
		ID:                       blobbackfill.WorkflowID,
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

	if _, err := s.temporalClient.ExecuteWorkflowInNamespace(ctx, blobBackfillNamespace, opts, blobbackfill.WorkflowName, blobbackfill.RangeRequest{
		Tables: tables,
	}); err != nil {
		return fmt.Errorf("unable to execute blob backfill workflow: %w", err)
	}

	return nil
}

type BackfillBlobsStatusResponse struct {
	Running        bool                   `json:"running"`
	Status         string                 `json:"status"`
	RunID          string                 `json:"run_id,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	ClosedAt       *time.Time             `json:"closed_at,omitempty"`
	ElapsedSeconds float64                `json:"elapsed_seconds,omitempty"`
	Elapsed        string                 `json:"elapsed,omitempty"`
	Progress       *blobbackfill.Progress `json:"progress,omitempty"`
}

// @ID						GetBackfillBlobsStatus
// @Summary				get dual-write blob backfill progress
// @Description			Reports whether the blob backfill workflow is still running and, when available, a live snapshot of how many rows per table have been mirrored to S3 so far.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	BackfillBlobsStatusResponse
// @Router					/v1/general/backfill-blobs [get]
func (s *service) GetBackfillBlobsStatus(ctx *gin.Context) {
	resp, err := s.getBlobBackfillStatus(ctx)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("blob backfill workflow has not been started")})
			return
		}
		ctx.Error(fmt.Errorf("unable to get blob backfill status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getBlobBackfillStatus(ctx context.Context) (*BackfillBlobsStatusResponse, error) {
	exec, err := s.temporalClient.DescribeWorkflowExecutionInNamespace(ctx, blobBackfillNamespace, blobbackfill.WorkflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe blob backfill workflow: %w", err)
	}

	info := exec.GetWorkflowExecutionInfo()
	status := info.GetStatus()
	resp := &BackfillBlobsStatusResponse{
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

	encoded, err := s.temporalClient.QueryWorkflowInNamespace(ctx, blobBackfillNamespace, blobbackfill.WorkflowID, "", blobbackfill.ProgressQueryType)
	if err != nil {
		// progress query is best-effort: a completed run on a closed workflow may not answer queries.
		return resp, nil
	}

	var progress blobbackfill.Progress
	if err := encoded.Get(&progress); err != nil {
		return nil, fmt.Errorf("unable to decode blob backfill progress: %w", err)
	}
	resp.Progress = &progress

	return resp, nil
}
