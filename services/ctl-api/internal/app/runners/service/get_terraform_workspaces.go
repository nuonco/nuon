package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetTerraformWorkspaces
// @Summary				get  terraform workspaces
// @Description.markdown	get_terraform_workspaces.md
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspace
// @Router					/v1/terraform-workspaces [get]
func (s *service) GetTerraformWorkpaces(ctx *gin.Context) {
	workspaces, err := s.listWorkspaces(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list workspaces: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workspaces)
}

func (s *service) listWorkspaces(ctx *gin.Context) ([]app.TerraformWorkspace, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := s.db.WithContext(ctx).Model(&app.TerraformWorkspace{}).Where("org_id = ?", orgID)
	if runner, err := cctx.RunnerFromContext(ctx); err == nil {
		if runner.RunnerGroup.Type != app.RunnerGroupTypeInstall || runner.RunnerGroup.OwnerID == "" {
			return nil, stderr.ErrNotFound{Err: fmt.Errorf("runner %s has no install", runner.ID), Description: "not found"}
		}
		query = query.Where(installWorkspaceScope(), installWorkspaceArgs(runner.RunnerGroup.OwnerID)...)
	}

	var workspaces []app.TerraformWorkspace
	if err = query.Find(&workspaces).Error; err != nil {
		return nil, err
	}

	return workspaces, nil
}

func installWorkspaceScope() string {
	return `(
		(owner_type IN ('installs', 'install') AND owner_id = ?)
		OR (owner_type = 'install_components' AND owner_id IN (SELECT id FROM install_components WHERE install_id = ? AND deleted_at = 0))
		OR (owner_type = 'install_sandbox_runs' AND owner_id IN (SELECT id FROM install_sandbox_runs WHERE install_id = ? AND deleted_at = 0))
		OR (owner_type = 'install_action_workflow_runs' AND owner_id IN (SELECT id FROM install_action_workflow_runs WHERE install_id = ? AND deleted_at = 0))
		OR (owner_type = 'install_sandboxes' AND owner_id IN (SELECT id FROM install_sandboxes WHERE install_id = ? AND deleted_at = 0))
		OR (owner_type = 'install_deploys' AND owner_id IN (
			SELECT install_deploys.id FROM install_deploys
			JOIN install_components ON install_components.id = install_deploys.install_component_id
			WHERE install_components.install_id = ? AND install_deploys.deleted_at = 0
		))
		OR (owner_type = 'install_workflow_steps' AND owner_id IN (
			SELECT install_workflow_steps.id FROM install_workflow_steps
			JOIN install_workflows ON install_workflows.id = install_workflow_steps.install_workflow_id
			WHERE install_workflows.owner_id = ? AND install_workflows.owner_type = 'installs' AND install_workflow_steps.deleted_at = 0
		))
	)`
}

func installWorkspaceArgs(installID string) []any {
	return []any{installID, installID, installID, installID, installID, installID, installID}
}
