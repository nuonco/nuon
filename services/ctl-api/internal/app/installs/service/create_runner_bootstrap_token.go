package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRunnerBootstrapTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// @ID						CreateRunnerBootstrapToken
// @Summary				create a bootstrap token for an install's runner
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	CreateRunnerBootstrapTokenResponse
// @Router					/v1/installs/{install_id}/runner-bootstrap-token [post]
func (s *service) CreateRunnerBootstrapToken(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	if install.RunnerID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("install %s has no runner", installID),
			Description: "install has no runner",
		})
		return
	}

	token, err := s.runnersHelpers.CreateBootstrapToken(ctx, install.RunnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create bootstrap token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateRunnerBootstrapTokenResponse{
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	})
}
