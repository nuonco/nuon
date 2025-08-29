package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallActionWorkflows
// @Summary				get an installs action workflows
// @Description.markdown	get_install_action_workflows.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/action-workflows [GET]
func (s *service) GetInstallActionWorkflows(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	installActionWorkflows, err := s.getInstallActionWorkflows(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installActionWorkflows)
}

func (s *service) getInstallActionWorkflows(ctx *gin.Context, installID string) ([]app.InstallActionWorkflow, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("InstallActionWorkflows", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("install_action_workflows.created_at DESC")
		}).
		Preload("InstallActionWorkflows.ActionWorkflow").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflows")
	}

	iaws, err := db.HandlePaginatedResponse(ctx, install.InstallActionWorkflows)
	if err != nil {
		return nil, errors.Wrap(err, "unable to handle paginated response")
	}

	install.InstallActionWorkflows = iaws

	return install.InstallActionWorkflows, nil
}
