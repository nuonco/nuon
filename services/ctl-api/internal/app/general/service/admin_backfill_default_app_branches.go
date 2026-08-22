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

	"github.com/nuonco/nuon/pkg/types/workflows/defaultappbranches"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const defaultAppBranchesNamespace = "general"

// AdminBackfillDefaultAppBranchesRequest optionally narrows the backfill to a
// subset of orgs, or previews the work without doing it.
type AdminBackfillDefaultAppBranchesRequest struct {
	OrgIDs []string `json:"org_ids"`
	DryRun bool     `json:"dry_run"`
}

// @ID						AdminBackfillDefaultAppBranches
// @Summary				backfill default app branches
// @Description			Starts a workflow that gives every app a branch named `default` with a single all-installs group, which is what `nuon apps sync` creates lazily on its first run once the org has default-app-branches enabled. Running this first means flipping the flag does not turn the next sync of every app into a migration.
// @Description
// @Description			Existing installs are recorded as members of that branch: every install the all-installs group resolves to and that no branch owns gets an active branch connection and its `app_branch_id` set, so the install reads as part of the branch instead of belonging to nothing. A branch holding installs only through an all-installs group is a weak owner, so a dev branch created later can still claim them by label or by ID. An app whose latest config on another branch already claims all installs is skipped, since two branches claiming everything would deploy over each other.
// @Description
// @Description			This does not touch the flag and does not trigger runs. An app that already has a `default` branch is still visited so its installs get connected.
// @Description
// @Description			Leave `org_ids` empty (or omit the body) to cover every org. Set `dry_run` to count the apps that would be backfilled without creating anything. Re-running drains whatever is left; an already-running backfill is reused rather than duplicated.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Param					request	body	AdminBackfillDefaultAppBranchesRequest	false	"optional org subset and dry run; empty processes every org"
// @Success				201	{string}	ok
// @Router					/v1/general/backfill-default-app-branches [post]
func (s *service) AdminBackfillDefaultAppBranches(ctx *gin.Context) {
	var req AdminBackfillDefaultAppBranchesRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(fmt.Errorf("unable to parse request: %w", err))
			return
		}
	}

	if err := s.startDefaultAppBranchesBackfill(ctx, req); err != nil {
		ctx.Error(fmt.Errorf("unable to start default app branch backfill: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) startDefaultAppBranchesBackfill(ctx context.Context, req AdminBackfillDefaultAppBranchesRequest) error {
	opts := tclient.StartWorkflowOptions{
		ID:                       defaultappbranches.WorkflowID,
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

	if _, err := s.temporalClient.ExecuteWorkflowInNamespace(ctx, defaultAppBranchesNamespace, opts, defaultappbranches.WorkflowName, defaultappbranches.Request{
		OrgIDs: req.OrgIDs,
		DryRun: req.DryRun,
	}); err != nil {
		return fmt.Errorf("unable to execute default app branch backfill workflow: %w", err)
	}

	return nil
}

type BackfillDefaultAppBranchesStatusResponse struct {
	Running        bool                         `json:"running"`
	Status         string                       `json:"status"`
	RunID          string                       `json:"run_id,omitempty"`
	StartedAt      *time.Time                   `json:"started_at,omitempty"`
	ClosedAt       *time.Time                   `json:"closed_at,omitempty"`
	ElapsedSeconds float64                      `json:"elapsed_seconds,omitempty"`
	Elapsed        string                       `json:"elapsed,omitempty"`
	Progress       *defaultappbranches.Progress `json:"progress,omitempty"`
}

// @ID						GetBackfillDefaultAppBranchesStatus
// @Summary				get default app branch backfill progress
// @Description			Reports whether the default app branch backfill is still running and, when available, how many apps were given a `default` branch, already had one, were skipped because another branch claims all installs, or failed, plus how many installs were connected to their branch.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	BackfillDefaultAppBranchesStatusResponse
// @Router					/v1/general/backfill-default-app-branches [get]
func (s *service) GetBackfillDefaultAppBranchesStatus(ctx *gin.Context) {
	resp, err := s.getDefaultAppBranchesBackfillStatus(ctx)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("default app branch backfill workflow has not been started")})
			return
		}
		ctx.Error(fmt.Errorf("unable to get default app branch backfill status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getDefaultAppBranchesBackfillStatus(ctx context.Context) (*BackfillDefaultAppBranchesStatusResponse, error) {
	exec, err := s.temporalClient.DescribeWorkflowExecutionInNamespace(ctx, defaultAppBranchesNamespace, defaultappbranches.WorkflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe default app branch backfill workflow: %w", err)
	}

	info := exec.GetWorkflowExecutionInfo()
	status := info.GetStatus()
	resp := &BackfillDefaultAppBranchesStatusResponse{
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

	encoded, err := s.temporalClient.QueryWorkflowInNamespace(ctx, defaultAppBranchesNamespace, defaultappbranches.WorkflowID, "", defaultappbranches.ProgressQueryType)
	if err != nil {
		// progress query is best-effort: a completed run on a closed workflow may not answer queries.
		return resp, nil
	}

	var progress defaultappbranches.Progress
	if err := encoded.Get(&progress); err != nil {
		return nil, fmt.Errorf("unable to decode default app branch backfill progress: %w", err)
	}
	resp.Progress = &progress

	return resp, nil
}
