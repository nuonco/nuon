package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AdminPromotionRequest struct {
	Tag string `json:"tag"`
}

// @ID						AdminPromotion
// @Summary				promotion callback.
// @Description.markdown	promotion.md
// @Param					req	body	AdminPromotionRequest	true	"Input"
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/general/promotion [POST]
func (s *service) AdminPromotion(ctx *gin.Context) {
	if _, err := cctx.AccountFromGinContext(ctx); err != nil {
		ctx.Error(stderr.ErrAuthentication{
			Err:         err,
			Description: "X-Nuon-Admin-Email header is required",
		})
		return
	}

	var req AdminPromotionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if _, err := s.adminOps.Promote(ctx, req.Tag); err != nil {
		ctx.Error(fmt.Errorf("unable to promote: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
