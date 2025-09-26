package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetCurrentAccount
// @Summary				Get current account
// @Description			Get the current account with user journeys and other data
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Success				200	{object}	app.Account
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/account [GET]
func (s *service) GetCurrentAccount(ctx *gin.Context) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Get full account with all relations
	fullAccount, err := s.getAccount(ctx, account.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, fullAccount)
}
