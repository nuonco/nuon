package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

type ExtendRunnerServiceTokenRequest struct {
	// defaults to one year
	Duration string `json:"duration" validate:"required" default:"8760h"`
}

// @ID						AdminExtendRunnerServiceAccount
// @Summary				extend a runner service account token
// @Description.markdown	extend_runner_service_account_token.md
// @Param					req	body	ExtendRunnerServiceTokenRequest true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Param					runner_id	path	string	true	"runner ID to fetch"
// @Produce				json
// @Success				200	{object}	app.RunnerGroup
// @Router					/v1/runners/{runner_id}/extend-service-account-token [POST]
func (s *service) AdminExtendRunnerServiceAccountToken(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req ExtendRunnerServiceTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse extend runner service account: %w", err))
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid time duration: %w", err))
		return
	}

	email := account.ServiceAccountEmail(runnerID)

	svcAcct, err := s.acctClient.FindAccount(ctx, email)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to find account"))
		return
	}

	if err := s.acctClient.ExtendToken(ctx, email, duration); err != nil {
		ctx.Error(errors.Wrap(err, "unable to extend token"))
		return
	}

	ctx.JSON(http.StatusOK, svcAcct)
}
