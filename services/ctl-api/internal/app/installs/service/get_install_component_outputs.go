package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						GetInstallComponentOutputs
//	@Summary				get an install component outputs
//	@Description.markdown	get_install_component_outputs.md
//	@Param					install_id		path	string	true	"install ID"
//	@Param					component_id	path	string	true	"component ID"
//	@Tags					installs
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{object}	map[string]interface{}
//	@Router					/v1/installs/{install_id}/components/{component_id}/outputs [get]
func (s *service) GetInstallComponentOutputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installCmp, err := s.getInstallComponentOutputs(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get outputs %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, installCmp)
}

func (s *service) getInstallComponentOutputs(ctx context.Context, installID, componentID string) (map[string]interface{}, error) {
	deploy, err := s.getInstallComponentLatestDeploy(ctx, installID, componentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component latest deploy")
	}

	var runnerJob app.RunnerJob
	res := s.db.WithContext(ctx).
		Where(app.RunnerJob{
			OwnerID: deploy.ID,
		}).
		Order("created_at DESC").
		First(&runnerJob)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job")
	}

	return runnerJob.ParsedOutputs, nil
}
