package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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
// @Description Exchanges the process runner token for a short-lived token that can write only this job. Succeeds only when the job status is available. A second claim returns 409 and does not return the existing token.
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

		acct, err := account.New(account.Params{DB: tx}).CreateServiceAccount(ctx.Request.Context(), job.ID, "runner-job")
		if err != nil {
			if isUniqueViolation(err) {
				return claimConflict{err: err}
			}
			return err
		}

		binding := app.ManagedServiceAccount{
			AccountID:   acct.ID,
			OrgID:       orgID,
			OwnerType:   plugins.TableName(tx, app.RunnerJob{}),
			OwnerID:     job.ID,
			Purpose:     app.ManagedServiceAccountPurposeRunnerJob,
			InstanceKey: "default",
		}
		if err := tx.Create(&binding).Error; err != nil {
			if isUniqueViolation(err) {
				return claimConflict{err: err}
			}
			return fmt.Errorf("bind job service account: %w", err)
		}

		roleCtx := cctx.SetAccountIDContext(ctx.Request.Context(), acct.ID)
		roleCtx = cctx.SetOrgIDContext(roleCtx, orgID)
		role := &app.Role{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeRunnerJob,
			Version:     app.RoleVersionV2,
			Title:       "Runner job",
			Description: "Scoped access to one runner job.",
			Policies: []app.Policy{{
				OrgID: generics.NewNullString(orgID),
				Name:  app.PolicyNameRunnerJob,
				Permissions: pgtype.Hstore{
					permissions.APIPermissionKey:                                 generics.ToPtr(string(permissions.APIRunner)),
					permissions.Object(orgID, permissions.KindRunnerJob, job.ID): permissions.PermissionAll.ToStrPtr(),
				},
			}},
		}
		if err := authz.New(authz.Params{DB: tx}).EnsureManagedServiceAccountRole(roleCtx, acct.ID, role); err != nil {
			return fmt.Errorf("create job role: %w", err)
		}

		now := time.Now()
		token := app.Token{
			CreatedByID: acct.ID,
			Token:       domains.NewUserTokenID(),
			TokenType:   app.TokenTypeNuon,
			ExpiresAt:   now.Add(jobTokenTTL),
			IssuedAt:    now,
			Issuer:      "nuon",
			AccountID:   acct.ID,
			OrgID:       orgID,
			SourceType:  app.TokenSourceTypeRunnerJob,
			SourceID:    job.ID,
		}
		if err := tx.Create(&token).Error; err != nil {
			return fmt.Errorf("create job token: %w", err)
		}
		if err := tx.Model(&app.RunnerJob{ID: job.ID}).Updates(app.RunnerJob{
			Status:            app.RunnerJobStatusInProgress,
			StatusDescription: "claimed",
		}).Error; err != nil {
			return fmt.Errorf("claim job: %w", err)
		}
		issued = &token
		return nil
	})
	return issued, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
