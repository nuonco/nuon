package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallActionWorkflowsLatestRuns
// @Summary					get latest runs for all action workflows by install id
// @Description.markdown	get_install_action_workflows_latest_run.md
// @Param					install_id	path			string	true	"install ID"
// @Param					trigger_types				query	string	false	"filter by action workflow trigger by types"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param		 			q							query	string	false	"search query for action workflow name"
// @Tags					actions
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/action-workflows/latest-runs [get]
func (s *service) GetInstallActionWorkflowsLatestRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	triggerTypes := ctx.Query("trigger_types")
	q := ctx.Query("q")
	var triggerTypesSlice []string
	if triggerTypes != "" {
		triggerTypesSlice = []string{triggerTypes}
	}

	iaws, err := s.getInstallActionWorkflowsLatestRun(ctx, org.ID, installID, triggerTypesSlice, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install action workflows: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, iaws)
}

func (s *service) getInstallActionWorkflowsLatestRun(ctx *gin.Context, orgID, installID string, triggerTypes []string, q string) ([]*app.InstallActionWorkflow, error) {
	iaws := []*app.InstallActionWorkflow{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("ActionWorkflow").
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			db = db.Scopes(
				scopes.WithOverrideTable("install_action_workflow_runs_latest_view_v1"),
			)
			return db
		}).
		Preload("Runs.RunnerJob", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithDisableViews)
		})

	if len(triggerTypes) > 0 {
		tx = tx.
			Joins("JOIN install_action_workflow_runs_latest_view_v1 ON install_action_workflows.id = install_action_workflow_runs_latest_view_v1.install_action_workflow_id").
			Where("install_action_workflow_runs_latest_view_v1.triggered_by_type IN ?", triggerTypes)
	}

	if q != "" {
		tx = tx.
			Joins("JOIN action_workflows ON install_action_workflows.action_workflow_id = action_workflows.id").
			Where("action_workflows.name ILIKE ?", "%"+q+"%")
	}

	res := tx.Find(&iaws, "install_action_workflows.org_id = ? AND install_action_workflows.install_id = ?", orgID, installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflows: %w", res.Error)
	}

	iaws, err := db.HandlePaginatedResponse(ctx, iaws)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return iaws, nil
}
