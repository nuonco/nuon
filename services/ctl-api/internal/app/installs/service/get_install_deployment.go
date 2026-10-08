package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID                    GetInstallDeployment
// @Summary               get a single normalized deployment for an install
// @Description.markdown  get_install_deployment.md
// @Param                 install_id   path  string  true  "install ID"
// @Param                 workflow_id  path  string  true  "workflow ID"
// @Tags                  installs
// @Accept                json
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400  {object}  stderr.ErrResponse
// @Failure               401  {object}  stderr.ErrResponse
// @Failure               403  {object}  stderr.ErrResponse
// @Failure               404  {object}  stderr.ErrResponse
// @Failure               500  {object}  stderr.ErrResponse
// @Success               200  {object}  InstallDeployment
// @Router                /v1/installs/{install_id}/deployments/{workflow_id} [GET]
func (s *service) GetInstallDeployment(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	workflowID := ctx.Param("workflow_id")

	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at ASC").Order("install_deploys.id ASC")
		}).
		Preload("InstallDeploys.InstallComponent").
		Preload("InstallDeploys.InstallComponent.Component").
		Preload("InstallDeploys.ComponentBuild").
		Where(app.Workflow{ID: workflowID, OwnerID: installID, OrgID: org.ID}).
		Where("plan_only = ?", false).
		First(&workflow).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install deployment: %w", err))
		return
	}

	deployments, err := s.buildInstallDeployments(ctx, org.ID, installID, []app.Workflow{workflow})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deployment: %w", err))
		return
	}
	if len(deployments) == 0 {
		ctx.Error(fmt.Errorf("workflow is not a deployment: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, deployments[0])
}
