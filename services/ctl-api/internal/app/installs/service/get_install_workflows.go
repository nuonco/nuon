package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallWorkflows
// @Summary				get an install workflows
// @Description.markdown	get_install_workflows.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
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
// @Success				200	{array}		app.InstallWorkflow
// @Router					/v1/installs/{install_id}/workflows [GET]
func (s *service) GetInstallWorkflows(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	installWorkflows, err := s.getInstallWorkflows(ctx, installID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflows"))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflows)
}

func (s *service) getInstallWorkflows(ctx *gin.Context, installID string) ([]app.InstallWorkflow, error) {
	var installWorkflows []app.InstallWorkflow
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("Steps").
		Where("install_id = ?", installID).
		Order("created_at desc").
		Find(&installWorkflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox runs: %w", res.Error)
	}

	installWorkflows, err := db.HandlePaginatedResponse(ctx, installWorkflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installWorkflows, nil
}
