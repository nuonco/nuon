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

	"github.com/nuonco/nuon/pkg/types/workflows/blobverify"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const blobVerifyNamespace = "general"

// AdminVerifyBlobsRequest optionally narrows the verification to a subset of
// dual-write tables. When empty, every supported table is processed.
type AdminVerifyBlobsRequest struct {
	Tables []string `json:"tables"`
}

// @ID						AdminVerifyBlobs
// @Summary				verify dual-write blobs against S3
// @Description			Starts an orchestrator workflow that enumerates the day-buckets of mirrored dual-write rows and verifies each day's rows against S3 blob storage: it confirms each blob's S3 object still matches its recorded checksum and that the blob content matches the canonical origin column. Re-running walks the full set again; an already-running verification is reused rather than duplicated.
// @Description
// @Description			Supported tables: install_states, install_workflow_step_approvals, runner_job_plans, runner_job_execution_outputs, terraform_workspace_states. Leave `tables` empty (or omit the body) to process all of them.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Param					request	body	AdminVerifyBlobsRequest	false	"optional table subset; empty processes all supported tables"
// @Success				201	{string}	ok
// @Router					/v1/general/verify-blobs [post]
func (s *service) AdminVerifyBlobs(ctx *gin.Context) {
	var req AdminVerifyBlobsRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(fmt.Errorf("unable to parse request: %w", err))
			return
		}
	}

	if err := s.startBlobVerify(ctx, req.Tables); err != nil {
		ctx.Error(fmt.Errorf("unable to start blob verify: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) startBlobVerify(ctx context.Context, tables []string) error {
	opts := tclient.StartWorkflowOptions{
		ID:                       blobverify.WorkflowID,
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

	if _, err := s.temporalClient.ExecuteWorkflowInNamespace(ctx, blobVerifyNamespace, opts, blobverify.WorkflowName, blobverify.RangeRequest{
		Tables: tables,
	}); err != nil {
		return fmt.Errorf("unable to execute blob verify workflow: %w", err)
	}

	return nil
}

type VerifyBlobsStatusResponse struct {
	Running        bool                 `json:"running"`
	Status         string               `json:"status"`
	RunID          string               `json:"run_id,omitempty"`
	StartedAt      *time.Time           `json:"started_at,omitempty"`
	ClosedAt       *time.Time           `json:"closed_at,omitempty"`
	ElapsedSeconds float64              `json:"elapsed_seconds,omitempty"`
	Elapsed        string               `json:"elapsed,omitempty"`
	Progress       *blobverify.Progress `json:"progress,omitempty"`
}

// @ID						GetVerifyBlobsStatus
// @Summary				get dual-write blob verify progress
// @Description			Reports whether the blob verify workflow is still running and, when available, a live snapshot of how many rows per table have been checked and how many mismatched their checksum or origin column.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	VerifyBlobsStatusResponse
// @Router					/v1/general/verify-blobs [get]
func (s *service) GetVerifyBlobsStatus(ctx *gin.Context) {
	resp, err := s.getBlobVerifyStatus(ctx)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("blob verify workflow has not been started")})
			return
		}
		ctx.Error(fmt.Errorf("unable to get blob verify status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getBlobVerifyStatus(ctx context.Context) (*VerifyBlobsStatusResponse, error) {
	exec, err := s.temporalClient.DescribeWorkflowExecutionInNamespace(ctx, blobVerifyNamespace, blobverify.WorkflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe blob verify workflow: %w", err)
	}

	info := exec.GetWorkflowExecutionInfo()
	status := info.GetStatus()
	resp := &VerifyBlobsStatusResponse{
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

	encoded, err := s.temporalClient.QueryWorkflowInNamespace(ctx, blobVerifyNamespace, blobverify.WorkflowID, "", blobverify.ProgressQueryType)
	if err != nil {
		// progress query is best-effort: a completed run on a closed workflow may not answer queries.
		return resp, nil
	}

	var progress blobverify.Progress
	if err := encoded.Get(&progress); err != nil {
		return nil, fmt.Errorf("unable to decode blob verify progress: %w", err)
	}
	resp.Progress = &progress

	return resp, nil
}
