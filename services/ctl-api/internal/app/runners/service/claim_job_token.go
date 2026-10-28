package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const jobTokenTTL = 6 * time.Hour

type claimConflict struct {
	err error
}

func (e claimConflict) Error() string {
	if e.err == nil {
		return "job is not available"
	}
	return e.err.Error()
}

func (e claimConflict) Unwrap() error { return e.err }

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
		var conflict claimConflict
		if errors.As(err, &conflict) || isUniqueViolation(err) {
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
	var issued *app.Token
	err := s.db.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var job app.RunnerJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.RunnerJob{ID: jobID, RunnerID: runnerID}).
			Take(&job).Error; err != nil {
			return err
		}
		if job.Status != app.RunnerJobStatusAvailable {
			return claimConflict{err: fmt.Errorf("job %s is %s", jobID, job.Status)}
		}

		jobAcct, err := account.New(account.Params{DB: tx}).FindAccount(ctx.Request.Context(), account.ServiceAccountEmail(job.ID))
		if err != nil {
			return fmt.Errorf("job service account not provisioned: %w", err)
		}

		now := time.Now()
		var existing int64
		if err := tx.Model(&app.Token{}).
			Where(app.Token{SourceType: app.TokenSourceTypeRunnerJob, SourceID: job.ID, AccountID: jobAcct.ID}).
			Where("expires_at > ?", now).
			Count(&existing).Error; err != nil {
			return fmt.Errorf("check existing job token: %w", err)
		}
		if existing > 0 {
			return claimConflict{err: fmt.Errorf("job %s already has a token", jobID)}
		}

		token := app.Token{
			CreatedByID: jobAcct.ID,
			Token:       domains.NewUserTokenID(),
			TokenType:   app.TokenTypeNuon,
			ExpiresAt:   now.Add(jobTokenTTL),
			IssuedAt:    now,
			Issuer:      "nuon",
			AccountID:   jobAcct.ID,
			OrgID:       orgID,
			SourceType:  app.TokenSourceTypeRunnerJob,
			SourceID:    job.ID,
		}
		if err := tx.Create(&token).Error; err != nil {
			return fmt.Errorf("create job token: %w", err)
		}
		// Job stays available until CreateJobExecution claims it (available → in_progress).
		issued = &token
		return nil
	})
	return issued, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
