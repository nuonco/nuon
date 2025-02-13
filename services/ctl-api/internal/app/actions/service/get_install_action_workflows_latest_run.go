package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)


// @ID GetInstallActionWorkflowsLatestRun
// @Summary	get latest runs for all action workflows by install id
// @Description.markdown	get_install_action_workflows_latest_run.md
// @Param			install_id	path	string	true	"install ID"
// @Tags			actions
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}	app.InstallActionWorkflow
// @Router			/v1/installs/{install_id}/action-workflows/latest-runs [get]
func (s *service) GetInstallActionWorkflowsLatestRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	iaws, err := s.getInstallActionWorkflowsLatestRun(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install action workflows: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, iaws)
}

func (s *service) getInstallActionWorkflowsLatestRun(ctx context.Context, orgID, installID string) ([]*app.InstallActionWorkflow, error) {
	iaws := []*app.InstallActionWorkflow{}
	res := s.db.WithContext(ctx).
		Preload("ActionWorkflow").
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_action_workflow_runs.created_at DESC").Limit(1)
		}).
		Find(&iaws, "org_id = ? AND install_id = ?", orgID, installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflows: %w", res.Error)
	}

	return iaws, nil
}
