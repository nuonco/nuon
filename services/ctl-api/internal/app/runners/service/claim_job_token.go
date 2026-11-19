package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ClaimRunnerJobTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// @ID ClaimRunnerJobToken
// @Summary Claim a job execution token
// @Description Exchanges the process runner token for a short-lived token for the job's pre-provisioned service account. Succeeds only when the job status is available. Does not mark the job in-progress — CreateJobExecution does that. A second claim returns 409 and does not return the existing token.
// @Tags runners/runner
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param runner_id path string true "runner ID"
// @Param job_id path string true "job ID"
// @Success 200 {object} ClaimRunnerJobTokenResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/runners/{runner_id}/jobs/{job_id}/token [POST]
func (s *service) ClaimRunnerJobToken(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	jobID := ctx.Param("job_id")
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if acct.Subject != runnerID {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("token subject %s cannot claim a job for runner %s", acct.Subject, runnerID),
			Description: "runner token does not match this runner",
		})
		return
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	token, err := s.claimRunnerJobToken(ctx, orgID, runnerID, jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: err, Description: "job not found"})
			return
		}
		if errors.Is(err, account.ErrJobClaimConflict) || isUniqueViolation(err) {
			ctx.Error(stderr.ErrConflict{
				Err:         err,
				Description: "job is not available",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to claim job token: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, ClaimRunnerJobTokenResponse{
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	})
}

func (s *service) claimRunnerJobToken(ctx *gin.Context, orgID, runnerID, jobID string) (*app.Token, error) {
	return s.acctClient.ClaimRunnerJobToken(ctx.Request.Context(), orgID, runnerID, jobID)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
