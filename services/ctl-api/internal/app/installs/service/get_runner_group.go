package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						GetInstallRunnerGroup
//	@Summary				Get an install's runner group
//	@Description.markdown	get_install_runner_group.md
//	@Tags					installs
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Param					install_id	path		string	true	"install ID"
//	@Failure				400			{object}	stderr.ErrResponse
//	@Failure				401			{object}	stderr.ErrResponse
//	@Failure				403			{object}	stderr.ErrResponse
//	@Failure				404			{object}	stderr.ErrResponse
//	@Failure				500			{object}	stderr.ErrResponse
//	@Success				200			{object}	app.RunnerGroup
//	@Router					/v1/installs/{install_id}/runner-group [GET]
func (s *service) GetInstallRunnerGroup(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	rg, err := s.getInstallRunnerGroup(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install runner group: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, rg)
}

func (s *service) getInstallRunnerGroup(ctx context.Context, installID string) (*app.RunnerGroup, error) {
	runnerGroup := app.RunnerGroup{}
	res := s.db.WithContext(ctx).
		Preload("Runners").
		Preload("Settings").
		Where(app.RunnerGroup{
			OwnerType: "installs",
			OwnerID:   installID,
		}).
		First(&runnerGroup)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install runner group %s: %w", installID, res.Error)
	}

	return &runnerGroup, nil
}
