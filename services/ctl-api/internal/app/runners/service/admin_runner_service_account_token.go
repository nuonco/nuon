package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

type AdminRunnerServiceAccountTokenRequest struct {
	// defaults to one year
	Duration string `json:"duration" validate:"required" default:"8760h"`

	Invalidate bool `json:"invalidate"`
}

type CreateTokenResponse struct {
	Token string `json:"token,omitzero"`
}

// @ID			AdminGetRunnerServiceAccountToken
// @BasePath	/v1/runners
// @Summary	Return a token for a runner service account
// @Schemes
// @Description.markdown	runner_service_account_token.md
// @Description			return all orgs
// @Param					runner_id	path	string									true	"install id"
// @Param					req			body	AdminRunnerServiceAccountTokenRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{object}	CreateTokenResponse
// @Router					/v1/runners/{runner_id}/service-account-token [POST]
func (s *service) AdminCreateRunnerServiceAccountToken(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminRunnerServiceAccountTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid time duration: %w", err))
		return
	}

	_, err = s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	email := account.ServiceAccountEmail(runnerID)

	if req.Invalidate {
		if err := s.acctClient.InvalidateTokens(ctx, email); err != nil {
			ctx.Error(err)
			return
		}
	}

	token, err := s.acctClient.CreateToken(ctx, email, duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateTokenResponse{
		Token: token.Token,
	})
}
