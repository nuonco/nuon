package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateInstallComponentDeployRequest struct {
	RequestID          string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	BuildID            string `json:"build_id"`
	DeployDependents   bool   `json:"deploy_dependents"`
	DeployDependencies bool   `json:"deploy_dependencies"`
	Role               string `json:"role,omitempty"`

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
// @Failure                 409 {object} stderr.ErrResponse
// @Failure                 500 {object} stderr.ErrResponse
// @Success                 201 {object} app.InstallDeploy
// @Router                  /v1/installs/{install_id}/components/{component_id}/deploys [post]
func (s *service) CreateInstallComponentDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")
	component, er := s.helpers.GetComponent(ctx, componentID)
	if er != nil {
		ctx.Error(fmt.Errorf("unable to get component %s: %w", componentID, er))
		return
	}

	if len(component.ComponentConfigs) > 0 {
		latestConfig := &component.ComponentConfigs[0]
		if latestConfig.IsToggleable() {
			// Enabled-state is the synthetic enabled install input; fall back to
			// the component's default_enabled when no value is set.
			enabled := latestConfig.GetDefaultEnabled()
			if ii, err := s.getLatestInstallInputs(ctx, installID); err == nil && ii != nil {
				if v, ok := ii.Values[config.EnabledOverrideInputName(component.Name)]; ok && v != nil {
					if parsed, perr := strconv.ParseBool(*v); perr == nil {
						enabled = parsed
					}
				}
			}
			if !enabled {
				ctx.Error(stderr.ErrUser{
					Err:         fmt.Errorf("component is disabled"),
					Description: "This component is disabled on this install. Enable it in your install config ([component_toggles] " + component.Name + " = true) or via the dashboard before deploying.",
				})
				return
			}
		}
	}

	var req CreateInstallDeployRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if req.RequestID != "" {
		deploy, err := s.createIdempotentInstallDeploy(ctx, installID, &req, componentID, component.Name)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, deploy)
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

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		installID,
		app.WorkflowTypeManualDeploy,
		map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: deploy.InstallComponent.Component.Name,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
			"deploy_dependencies":                     strconv.FormatBool(req.DeployDependencies),
		},
		req.PlanOnly,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UpdateDeployWithWorkflowID(ctx, deploy.ID, workflow.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install deploy with workflow ID: %w", err))
		return
	}

	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	deploy.WorkflowID = &workflow.ID
	ctx.JSON(http.StatusCreated, deploy)
}

type CreateInstallDeployRequest struct {
	RequestID          string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	BuildID            string `json:"build_id"`
	DeployDependents   bool   `json:"deploy_dependents"`
	DeployDependencies bool   `json:"deploy_dependencies"`
	Role               string `json:"role,omitempty"`

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
// @Failure                 409 {object} stderr.ErrResponse
// @Failure                 500 {object} stderr.ErrResponse
// @Success                 201 {object} app.InstallDeploy
// @Router                  /v1/installs/{install_id}/deploys [post]
func (s *service) CreateInstallDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallDeployRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if req.RequestID != "" {
		deploy, err := s.createIdempotentInstallDeploy(ctx, installID, &req, "", "")
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, deploy)
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

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		installID,
		app.WorkflowTypeManualDeploy,
		map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: deploy.InstallComponent.Component.Name,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
			"deploy_dependencies":                     strconv.FormatBool(req.DeployDependencies),
		},
		req.PlanOnly,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UpdateDeployWithWorkflowID(ctx, deploy.ID, workflow.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install deploy with workflow ID: %w", err))
		return
	}

	queueID2, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID2, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	deploy.WorkflowID = &workflow.ID
	ctx.JSON(http.StatusCreated, deploy)
}

func (s *service) createIdempotentInstallDeploy(
	ctx context.Context,
	installID string,
	req *CreateInstallDeployRequest,
	pathComponentID string,
	componentNameHint string,
) (*app.InstallDeploy, error) {
	hashReq := *req
	hashReq.RequestID = ""
	var hash string
	var err error
	if pathComponentID != "" {
		hash, err = request.Hash(struct {
			CreateInstallDeployRequest
			ComponentID string `json:"component_id"`
		}{
			CreateInstallDeployRequest: hashReq,
			ComponentID:                pathComponentID,
		})
	} else {
		hash, err = request.Hash(hashReq)
	}
	if err != nil {
		return nil, err
	}

	var createdDeploy *app.InstallDeploy
	var createdComponentID string
	var createdInstallComponentID string
	workflow, created, err := s.helpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
		InstallID:    installID,
		WorkflowType: app.WorkflowTypeManualDeploy,
		PlanOnly:     req.PlanOnly,
		Role:         req.Role,
		RequestID:    req.RequestID,
		RequestHash:  hash,
		Operation:    "manual-deploy",
		QueueName:    installhelpers.InstallWorkflowsQueueName,
	}, &installhelpers.IdempotentInstallWorkflowHooks{
		Prepare: func(tx *gorm.DB) (map[string]string, error) {
			deploy, componentID, componentName, installComponentID, err := s.insertInstallDeploy(ctx, tx, installID, req)
			if err != nil {
				return nil, err
			}
			createdDeploy = deploy
			createdComponentID = componentID
			createdInstallComponentID = installComponentID
			nameSuffix := componentNameHint
			if nameSuffix == "" {
				nameSuffix = componentName
			}
			return map[string]string{
				app.WorkflowMetadataKeyWorkflowNameSuffix: nameSuffix,
				"install_deploy_id":                       deploy.ID,
				"deploy_dependents":                       strconv.FormatBool(req.DeployDependents),
				"deploy_dependencies":                     strconv.FormatBool(req.DeployDependencies),
			}, nil
		},
		AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error {
			if createdDeploy == nil {
				return fmt.Errorf("install deploy was not created")
			}
			return tx.WithContext(ctx).Model(&app.InstallDeploy{}).
				Where("id = ?", createdDeploy.ID).
				Update("install_workflow_id", wf.ID).Error
		},
	})
	if err != nil {
		return nil, err
	}

	if created && createdDeploy != nil {
		s.emitInstallDeployAudit(ctx, installID, createdDeploy.ID, req.BuildID, createdComponentID, createdInstallComponentID, createdDeploy.Type)
	}

	deployID := ""
	if createdDeploy != nil {
		deployID = createdDeploy.ID
	} else {
		deployID = generics.FromPtrStr(workflow.Metadata["install_deploy_id"])
	}

	deploy, err := s.getInstallDeploy(ctx, installID, deployID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install deploy %s: %w", deployID, err)
	}
	deploy.WorkflowID = &workflow.ID
	return deploy, nil
}

func (s *service) createInstallDeploy(ctx context.Context, installID string, req *CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	deploy, componentID, _, installComponentID, err := s.insertInstallDeploy(ctx, s.db, installID, req)
	if err != nil {
		return nil, err
	}
	s.emitInstallDeployAudit(ctx, installID, deploy.ID, req.BuildID, componentID, installComponentID, deploy.Type)
	return deploy, nil
}

func (s *service) insertInstallDeploy(
	ctx context.Context,
	db *gorm.DB,
	installID string,
	req *CreateInstallDeployRequest,
) (*app.InstallDeploy, string, string, string, error) {
	var build app.ComponentBuild
	res := db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		First(&build, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, "", "", "", fmt.Errorf("unable to get build %s: %w", req.BuildID, res.Error)
	}

	componentID := build.ComponentConfigConnection.ComponentID
	componentName := build.ComponentConfigConnection.Component.Name

	var installCmp app.InstallComponent
	res = db.WithContext(ctx).Where(app.InstallComponent{
		InstallID:   installID,
		ComponentID: componentID,
	}).First(&installCmp)
	if res.Error != nil {
		return nil, "", "", "", fmt.Errorf("unable to create install component: %w", res.Error)
	}

	typ := app.InstallDeployTypeApply
	deploy := app.InstallDeploy{
		Status:             "queued",
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   req.BuildID,
		InstallComponentID: installCmp.ID,
		Type:               typ,
		Role:               req.Role,
	}

	res = db.WithContext(ctx).Create(&deploy)
	if res.Error != nil {
		return nil, "", "", "", fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	return &deploy, componentID, componentName, installCmp.ID, nil
}

func (s *service) emitInstallDeployAudit(
	ctx context.Context,
	installID, deployID, buildID, componentID, installComponentID string,
	typ app.InstallDeployType,
) {
	s.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallDeploy,
		Message:     "install deploy created",
		Outcome:     audit.OutcomeStarted,
		InstallID:   installID,
		ComponentID: componentID,
		SubjectID:   deployID,
		SubjectType: "install_deploys",
		Attrs: map[string]string{
			"deploy.id":            deployID,
			"deploy.type":          string(typ),
			"component_build.id":   buildID,
			"install_component.id": installComponentID,
		},
	})
}
