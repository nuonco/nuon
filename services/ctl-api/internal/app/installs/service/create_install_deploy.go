package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateInstallComponentDeployRequest struct {
	BuildID          string `json:"build_id"`
	DeployDependents bool   `json:"deploy_dependents"`

	PlanOnly bool `json:"plan_only"`
}

func (c *CreateInstallComponentDeployRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID                      CreateInstallComponentDeploy
// @Summary                 deploy a build to an install
// @Description.markdown    create_install_deploy.md
// @Param                   install_id  path    string                      true    "install ID"
// @Param                   component_id path   string                      true    "component ID"
// @Param                   req         body    CreateInstallComponentDeployRequest  true    "Input"
// @Tags                    installs
// @Accept                  json
// @Produce                 json
// @Security                APIKey
// @Security                OrgID
// @Failure                 400 {object} stderr.ErrResponse
// @Failure                 401 {object} stderr.ErrResponse
// @Failure                 403 {object} stderr.ErrResponse
// @Failure                 404 {object} stderr.ErrResponse
// @Failure                 500 {object} stderr.ErrResponse
// @Success                 201 {object} app.InstallDeploy
// @Router                  /v1/installs/{install_id}/components/{component_id}/deploys [post]
func (s *service) CreateInstallComponentDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")
	_, er := s.helpers.GetComponent(ctx, componentID)
	if er != nil {
		ctx.Error(fmt.Errorf("unable to get component %s: %w", componentID, er))
		return
	}
	var req CreateInstallDeployRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	deploy, err := s.createInstallDeploy(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	deploy, err = s.getInstallDeploy(ctx, installID, deploy.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get newly created deploy %s:  %w", deploy.ID, err))
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		installID,
		app.WorkflowTypeManualDeploy,
		map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: deploy.InstallComponent.Component.Name,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
		},
		app.StepErrorBehaviorAbort,
		req.PlanOnly,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UpdateDeployWithWorkflowID(ctx, deploy.ID, workflow.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install deploy with workflow ID: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, deploy)
}

type CreateInstallDeployRequest struct {
	BuildID          string `json:"build_id"`
	DeployDependents bool   `json:"deploy_dependents"`

	PlanOnly bool `json:"plan_only"`
}

func (c *CreateInstallDeployRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID                      CreateInstallDeploy
// @Summary                 deploy a build to an install
// @Description.markdown    create_install_deploy.md
// @Param                   install_id  path    string                      true    "install ID"
// @Param                   req         body    CreateInstallDeployRequest  true    "Input"
// @Tags                    installs
// @Accept                  json
// @Produce                 json
// @Security                APIKey
// @Security                OrgID
// @Deprecated              true
// @Failure                 400 {object} stderr.ErrResponse
// @Failure                 401 {object} stderr.ErrResponse
// @Failure                 403 {object} stderr.ErrResponse
// @Failure                 404 {object} stderr.ErrResponse
// @Failure                 500 {object} stderr.ErrResponse
// @Success                 201 {object} app.InstallDeploy
// @Router                  /v1/installs/{install_id}/deploys [post]
func (s *service) CreateInstallDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallDeployRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	deploy, err := s.createInstallDeploy(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	deploy, err = s.getInstallDeploy(ctx, installID, deploy.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get newly created deploy %s:  %w", deploy.ID, err))
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		installID,
		app.WorkflowTypeManualDeploy,
		map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: deploy.InstallComponent.Component.Name,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
		},
		app.StepErrorBehaviorAbort,
		req.PlanOnly,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UpdateDeployWithWorkflowID(ctx, deploy.ID, workflow.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install deploy with workflow ID: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, deploy)
}

func (s *service) createInstallDeploy(ctx context.Context, installID string, req *CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	var build app.ComponentBuild
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		First(&build, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build %s: %w", req.BuildID, res.Error)
	}

	// ensure that the install component exists
	var install app.Install
	res = s.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Where("component_id = ?", build.ComponentConfigConnection.ComponentID).
				Where("install_id = ?", installID)
		}).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	// if the install component does not exist, create it.
	if len(install.InstallComponents) != 1 {
		err := s.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			First(&install, "id = ?", installID).
			Association("InstallComponents").
			Append(&app.InstallComponent{
				ComponentID: build.ComponentConfigConnection.ComponentID,
			})
		if err != nil {
			return nil, fmt.Errorf("unable to create missing install component: %w", err)
		}
	}

	// create deploy
	var installCmp app.InstallComponent
	res = s.db.WithContext(ctx).Where(app.InstallComponent{
		InstallID:   installID,
		ComponentID: build.ComponentConfigConnection.ComponentID,
	}).First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install component: %w", res.Error)
	}

	typ := app.InstallDeployTypeApply
	deploy := app.InstallDeploy{
		Status:             "queued",
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   req.BuildID,
		InstallComponentID: installCmp.ID,
		Type:               typ,
	}

	res = s.db.WithContext(ctx).Create(&deploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	return &deploy, nil
}
