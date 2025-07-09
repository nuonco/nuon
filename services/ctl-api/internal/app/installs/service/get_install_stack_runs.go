package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallStackRuns
// @Summary				get an install's stack runs
// @Description	get install stack runs
// @Param install_id					path	string	true "install ID"
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
// @Success				200	{object}		app.InstallStackVersionRun
// @Router					/v1/installs/{install_id}/stack-runs [get]
func (s *service) GetInstallStackRuns(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	runs, err := s.getInstallLatestStackRunsByStackID(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runs)
}

func (s *service) getInstallLatestStackRunsByStackID(ctx context.Context, installID string) ([]app.InstallStackVersionRun, error) {
	var runs []app.InstallStackVersionRun

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).

		// join component-releases to component-builds to component-config-connections to components
		Joins("JOIN install_stack_versions ON install_stack_versions.id=install_stack_version_runs.install_stack_version_id").
		Joins("JOIN install_stacks ON install_stacks.id=install_stack_versions.install_stack_id").
		Where("install_stacks.install_id = ?", installID).
		Order("install_stack_version_runs.created_at DESC").
		Find(&runs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component releases")
	}

	return runs, nil
}
