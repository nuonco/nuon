package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/revokeunusedtokens"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const revokeUnusedTokensBatchSize = 100

type AdminRevokeUnusedTokensRequest struct {
	OrgID string `json:"org_id,omitempty"`
}

type AdminRevokeUnusedTokensResponse struct {
	EnqueuedCount int `json:"enqueued_count"`
	FailedCount   int `json:"failed_count"`
}

// @ID						AdminRevokeRunnerUnusedTokens
// @Summary				revoke tokens no live process of a runner is using
// @Description.markdown	revoke_unused_tokens.md
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/revoke-unused-tokens [POST]
func (s *service) AdminRevokeRunnerUnusedTokens(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &revokeunusedtokens.Signal{RunnerID: runnerID}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue revoke-unused-tokens signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, true)
}

// @ID						AdminRevokeUnusedTokens
// @Summary				revoke tokens no live process is using, for runners in default orgs
// @Description.markdown	revoke_unused_tokens_all.md
// @Param					req	body	AdminRevokeUnusedTokensRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{object}	AdminRevokeUnusedTokensResponse
// @Router					/v1/runners/revoke-unused-tokens [POST]
func (s *service) AdminRevokeUnusedTokens(ctx *gin.Context) {
	var req AdminRevokeUnusedTokensRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	resp, err := s.enqueueRevokeUnusedTokens(ctx, req.OrgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue revoke-unused-tokens signals: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (s *service) enqueueRevokeUnusedTokens(ctx context.Context, orgID string) (*AdminRevokeUnusedTokensResponse, error) {
	resp := &AdminRevokeUnusedTokensResponse{}
	lastID := ""

	for {
		var runnerIDs []string
		query := s.db.WithContext(ctx).
			Model(&app.Runner{}).
			Joins("JOIN orgs ON runners.org_id = orgs.id AND orgs.org_type = ?", app.OrgTypeDefault).
			Where("EXISTS (SELECT 1 FROM runner_processes rp WHERE rp.runner_id = runners.id AND rp.deleted_at = 0 AND rp.composite_status->>'status' IN ?)",
				helpers.TokenHoldingRunnerProcessStatuses()).
			Where("runners.id > ?", lastID).
			Order("runners.id ASC").
			Limit(revokeUnusedTokensBatchSize)
		if orgID != "" {
			query = query.Where(app.Runner{OrgID: orgID})
		}
		if err := query.Pluck("runners.id", &runnerIDs).Error; err != nil {
			return nil, fmt.Errorf("unable to list runners: %w", err)
		}
		if len(runnerIDs) == 0 {
			return resp, nil
		}

		for _, runnerID := range runnerIDs {
			if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &revokeunusedtokens.Signal{RunnerID: runnerID}); err != nil {
				s.l.Error("unable to enqueue revoke-unused-tokens signal", zap.String("runner_id", runnerID), zap.Error(err))
				resp.FailedCount++
				continue
			}
			resp.EnqueuedCount++
		}

		lastID = runnerIDs[len(runnerIDs)-1]
	}
}
