package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallActionWorkflowRuns
// @Summary				get action workflow runs by install id
// @Description.markdown	get_install_action_workflow_runs.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallActionWorkflowRun
// @Router					/v1/installs/{install_id}/action-workflows/runs [get]
func (s *service) GetInstallActionWorkflowRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	runs, err := s.findInstallActionWorkflowRuns(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install action workflow runs %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, runs)
}

func (s *service) findInstallActionWorkflowRuns(ctx *gin.Context, orgID, installID string) ([]*app.InstallActionWorkflowRun, error) {
	runs := []*app.InstallActionWorkflowRun{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where("org_id = ? AND install_id = ?", orgID, installID).
		Order("created_at desc").
		Find(&runs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflow runs: %w", res.Error)
	}

	runs, err := db.HandlePaginatedResponse(ctx, runs)
	if err != nil {
		return nil, fmt.Errorf("unable to get install action workflow runs: %w", err)
	}

	return runs, nil
}
