package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type PruneTokensResponse struct {
	InvalidatedCount int64 `json:"invalidated_count"`
}

// @ID						PruneTokens
// @Summary				Prune old tokens for a runner
// @Description.markdown	prune_tokens.md
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					runner_id	path		string	true	"runner ID"
// @Failure				400			{object}	stderr.ErrResponse
// @Failure				401			{object}	stderr.ErrResponse
// @Failure				403			{object}	stderr.ErrResponse
// @Failure				404			{object}	stderr.ErrResponse
// @Failure				500			{object}	stderr.ErrResponse
// @Success				200			{object}	PruneTokensResponse
// @Router					/v1/runners/{runner_id}/prune-tokens [POST]
func (s *service) PruneTokens(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	// Verify runner belongs to caller's org
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runner, err := s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil || runner == nil {
		ctx.Error(fmt.Errorf("unable to get runner %s: %w", runnerID, err))
		return
	}

	count, err := s.helpers.InvalidateOldTokens(ctx, runner.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to prune tokens for runner %s: %w", runner.ID, err))
		return
	}

	ctx.JSON(http.StatusOK, PruneTokensResponse{
		InvalidatedCount: count,
	})
}
