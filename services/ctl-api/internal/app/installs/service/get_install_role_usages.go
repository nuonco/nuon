package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallRoleUsages
// @Summary				get install role usages
// @Description			get workflows that used a particular role, filtered by unrendered role name
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Param					role_name	query	string	true	"unrendered role name template"
// @Param					offset		query	int		false	"offset of results to return"	Default(0)
// @Param					limit		query	int		false	"limit of results to return"	Default(10)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallRoleUsage
// @Router					/v1/installs/{install_id}/roles/usages [get]
func (s *service) GetInstallRoleUsages(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	roleName := ctx.Query("role_name")

	if roleName == "" {
		ctx.Error(fmt.Errorf("role_name query parameter is required"))
		return
	}

	var installRoleIDs []string
	res := s.db.WithContext(ctx).Unscoped().
		Model(&app.InstallRoles{}).
		Joins("JOIN app_awsiam_role_configs arc ON arc.id = install_roles.app_role_config_id").
		Where("install_roles.install_id = ? AND install_roles.org_id = ?", installID, org.ID).
		Where("arc.name = ?", roleName).
		Pluck("install_roles.id", &installRoleIDs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find install roles: %w", res.Error))
		return
	}

	if len(installRoleIDs) == 0 {
		ctx.JSON(http.StatusOK, []app.InstallRoleUsage{})
		return
	}

	var usages []app.InstallRoleUsage
	res = s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("RunnerJob").
		Where("install_role_id IN ? AND org_id = ?", installRoleIDs, org.ID).
		Order("created_at DESC").
		Find(&usages)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install role usages: %w", res.Error))
		return
	}

	usages, err = db.HandlePaginatedResponse(ctx, usages)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	for i := range usages {
		if usages[i].RunnerJob.OwnerID == "" {
			continue
		}

		var step app.WorkflowStep
		res := s.db.WithContext(ctx).
			Where(app.WorkflowStep{
				StepTargetID:   usages[i].RunnerJob.OwnerID,
				StepTargetType: usages[i].RunnerJob.OwnerType,
			}).
			First(&step)
		if res.Error != nil {
			continue
		}

		usages[i].WorkflowStepID = step.ID

		var workflow app.Workflow
		if err := s.db.WithContext(ctx).
			Where("id = ?", step.InstallWorkflowID).
			First(&workflow).Error; err == nil {
			usages[i].Workflow = &workflow
		}
	}

	ctx.JSON(http.StatusOK, usages)
}
