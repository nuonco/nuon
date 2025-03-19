package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

type AdminPromotionRequest struct {
	Tag string `json:"tag"`
}

// @ID AdminPromotion
// @Summary	promotion callback.
// @Description.markdown promotion.md
// @Param			req	body	AdminPromotionRequest	true	"Input"
// @Tags general/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		201	{string} ok
// @Router			/v1/general/promotion [POST]
func (s *service) AdminPromotion(ctx *gin.Context) {
	var req AdminPromotionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to promote"))
		return
	}

	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationRestart,
		Tag:  req.Tag,
	})
	s.evClient.Send(ctx, "general", &signals.Signal{
		Type: signals.OperationPromotion,
		Tag:  req.Tag,
	})

	ctx.JSON(http.StatusCreated, "ok")
}
