package service

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type AdminInvalidateRunnerServiceAccountTokenRequest struct{}

// @ID			AdminInvalidateRunnerServiceAccountToken
// @BasePath	/v1/runners
// @Summary	Invalidate a token for a runner service account
// @Schemes
// @Description.markdown invalidate_runner_service_account_token.md
// @Description			return all orgs
// @Param					runner_id	path	string									true	"install id"
// @Param req			body AdminInvalidateRunnerServiceAccountTokenRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/runners/{runner_id}/invalidate-service-account-token [POST]
func (s *service) AdminInvalidateRunnerServiceAccountToken(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminInvalidateRunnerServiceAccountTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	email := account.ServiceAccountEmail(runnerID)
	if err := s.acctClient.InvalidateTokens(ctx, email); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
