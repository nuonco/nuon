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

	"github.com/nuonco/nuon/pkg/types/workflows/phonehomesecretbackfill"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// phoneHomeSecretBackfillNamespace is "installs", not "general": the reconciler
// activity the workflow drives is registered on the installs worker, and a workflow
// can only execute activities registered in its own namespace.
const phoneHomeSecretBackfillNamespace = "installs"

// @ID						AdminBackfillPhoneHomeSecrets
// @Summary				backfill install phone home secrets
// @Description			Starts a durable workflow that provisions phone-home credentials for existing installs in orgs with the phone-home-auth feature enabled: minting a token per live stack version, publishing the phone_home_id-to-token map to Secrets Manager in the management account, and granting the install's phone-home role cross-account read. Idempotent, and re-triggering reuses an in-progress backfill rather than starting a duplicate. Note this is pre-provisioning only — an already-deployed phone-home Lambda does not know the secret's location and will not send a token until its stack version is regenerated and re-applied.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				202	{string}	ok
// @Router					/v1/general/backfill-phone-home-secrets [post]
func (s *service) AdminBackfillPhoneHomeSecrets(ctx *gin.Context) {
	if err := s.startPhoneHomeSecretBackfill(ctx); err != nil {
		ctx.Error(fmt.Errorf("unable to start phone home secret backfill: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{"status": "ok"})
}

func (s *service) startPhoneHomeSecretBackfill(ctx context.Context) error {
	opts := tclient.StartWorkflowOptions{
		ID:                       phonehomesecretbackfill.WorkflowID,
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

	if _, err := s.temporalClient.ExecuteWorkflowInNamespace(ctx, phoneHomeSecretBackfillNamespace, opts,
		phonehomesecretbackfill.WorkflowName, phonehomesecretbackfill.Request{}); err != nil {
		return fmt.Errorf("unable to execute phone home secret backfill workflow: %w", err)
	}

	return nil
}

type BackfillPhoneHomeSecretsStatusResponse struct {
	Running        bool                              `json:"running"`
	Status         string                            `json:"status"`
	RunID          string                            `json:"run_id,omitempty"`
	StartedAt      *time.Time                        `json:"started_at,omitempty"`
	ClosedAt       *time.Time                        `json:"closed_at,omitempty"`
	ElapsedSeconds float64                           `json:"elapsed_seconds,omitempty"`
	Elapsed        string                            `json:"elapsed,omitempty"`
	Progress       *phonehomesecretbackfill.Progress `json:"progress,omitempty"`
}

// @ID						GetBackfillPhoneHomeSecretsStatus
// @Summary				get install phone home secret backfill progress
// @Description			Reports whether the phone-home secret backfill workflow is still running and, when available, a live snapshot of how many installs have been processed, how many secrets were ensured, how many tokens were minted, and how many installs were skipped by the eligibility gate.
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				200	{object}	BackfillPhoneHomeSecretsStatusResponse
// @Router					/v1/general/backfill-phone-home-secrets [get]
func (s *service) GetBackfillPhoneHomeSecretsStatus(ctx *gin.Context) {
	resp, err := s.getPhoneHomeSecretBackfillStatus(ctx)
	if err != nil {
		var notFoundErr *serviceerror.NotFound
		if errors.As(err, &notFoundErr) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("phone home secret backfill workflow has not been started")})
			return
		}
		ctx.Error(fmt.Errorf("unable to get phone home secret backfill status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getPhoneHomeSecretBackfillStatus(ctx context.Context) (*BackfillPhoneHomeSecretsStatusResponse, error) {
	exec, err := s.temporalClient.DescribeWorkflowExecutionInNamespace(ctx, phoneHomeSecretBackfillNamespace, phonehomesecretbackfill.WorkflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe phone home secret backfill workflow: %w", err)
	}

	info := exec.GetWorkflowExecutionInfo()
	status := info.GetStatus()
	resp := &BackfillPhoneHomeSecretsStatusResponse{
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

	encoded, err := s.temporalClient.QueryWorkflowInNamespace(ctx, phoneHomeSecretBackfillNamespace, phonehomesecretbackfill.WorkflowID, "", phonehomesecretbackfill.ProgressQueryType)
	if err != nil {
		// progress query is best-effort: a completed run on a closed workflow may not answer queries.
		return resp, nil
	}

	var progress phonehomesecretbackfill.Progress
	if err := encoded.Get(&progress); err != nil {
		return nil, fmt.Errorf("unable to decode phone home secret backfill progress: %w", err)
	}
	resp.Progress = &progress

	return resp, nil
}
