package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PruneRunnerTokensResponse struct {
	InvalidatedCount int64 `json:"invalidated_count"`
}

// @ID						PruneInstallRunnerTokens
// @Summary				Prune old tokens for an install's runner
// @Description.markdown	prune_runner_tokens.md
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					install_id	path		string	true	"install ID"
// @Failure				400			{object}	stderr.ErrResponse
// @Failure				401			{object}	stderr.ErrResponse
// @Failure				403			{object}	stderr.ErrResponse
// @Failure				404			{object}	stderr.ErrResponse
// @Failure				500			{object}	stderr.ErrResponse
// @Success				200			{object}	PruneRunnerTokensResponse
// @Router					/v1/installs/{install_id}/prune-runner-tokens [POST]
func (s *service) PruneInstallRunnerTokens(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	rg, err := s.getInstallRunnerGroup(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install runner group: %w", err))
		return
	}

	if len(rg.Runners) == 0 {
		ctx.Error(fmt.Errorf("install has no runners"))
		return
	}

	var totalInvalidated int64
	for _, runner := range rg.Runners {
		count, err := s.runnersHelpers.InvalidateOldTokens(ctx, runner.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to prune tokens for runner %s: %w", runner.ID, err))
			return
		}
		totalInvalidated += count
	}

	ctx.JSON(http.StatusOK, PruneRunnerTokensResponse{
		InvalidatedCount: totalInvalidated,
	})
}
