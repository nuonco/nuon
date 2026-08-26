package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type RecoverInstallComponentHelmReleaseRequest struct {
	Role string `json:"role,omitempty"`
}

func (c *RecoverInstallComponentHelmReleaseRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						RecoverInstallComponentHelmRelease
// @Summary				recover a stuck helm release for an install component
// @Description.markdown	recover_install_component_helm_release.md
// @Param					install_id		path	string										true	"install ID"
// @Param					component_id	path	string										true	"component ID"
// @Param					req				body	RecoverInstallComponentHelmReleaseRequest	false	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id}/components/{component_id}/recover-helm-release [post]
func (s *service) RecoverInstallComponentHelmRelease(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	var req RecoverInstallComponentHelmReleaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.helpers.GetInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	component, err := s.helpers.GetComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	if component.Type != app.ComponentTypeHelmChart {
		ctx.Error(stderr.ErrUser{
			Err: errors.New("component is not a helm chart"),
			Description: "Only Helm chart components have a Helm release to recover. " +
				component.Name + " is a " + string(component.Type) + " component.",
		})
		return
	}

	installComponent, err := s.helpers.GetInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component: %w", err))
		return
	}

	// The plan needs a build to resolve the release name, namespace and driver.
	deploy, err := s.getLatestDeployForRecovery(ctx, installComponent.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Gated on live runner jobs, not workflow status: a died deploy parks its
	// workflow in failed-pending-retry forever, which is this action's whole case.
	running, err := s.hasRunningDeployJob(ctx, installComponent.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if running {
		ctx.Error(stderr.ErrConflict{
			Err: errors.New("a deploy is already running for this component"),
			Description: "A job is currently running for " + component.Name + ". Wait for it to finish, or cancel it, " +
				"before recovering the Helm release — recovering while Helm is mid-operation can corrupt the release.",
		})
		return
	}

	recoveryDeploy, err := s.createRecoveryDeploy(ctx, installComponent.ID, deploy.ComponentBuildID, component.ID, installID, req.Role)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create recovery deploy: %w", err))
		return
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		install.ID,
		app.WorkflowTypeRecoverHelmRelease,
		map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: component.Name,
			"component_id":      component.ID,
			"install_deploy_id": recoveryDeploy.ID,
		},
		false,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.helpers.UpdateDeployWithWorkflowID(ctx, recoveryDeploy.ID, workflow.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to update install deploy with workflow ID: %w", err))
		return
	}

	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}

// getLatestDeployForRecovery returns the deploy whose build describes the release
// to recover. A component that has never been deployed has no release, and a
// caller who reaches here has misdiagnosed the problem.
func (s *service) getLatestDeployForRecovery(ctx context.Context, installComponentID string) (*app.InstallDeploy, error) {
	var deploy app.InstallDeploy
	res := s.db.WithContext(ctx).
		Where(app.InstallDeploy{InstallComponentID: installComponentID}).
		Where("component_build_id IS NOT NULL AND component_build_id != ''").
		Where("install_deploys.type != ?", app.InstallDeployTypeRecover).
		Order("install_deploys.created_at DESC").
		First(&deploy)
	if res.Error != nil {
		return nil, stderr.ErrConflict{
			Err: errors.New("component has never been deployed"),
			Description: "This component has no deploy history on this install, so there is no Helm release to recover. " +
				"Deploy the component instead.",
		}
	}

	return &deploy, nil
}

// hasRunningDeployJob reports whether any deploy of this install component still
// has a runner job that has not reached a terminal status.
func (s *service) hasRunningDeployJob(ctx context.Context, installComponentID string) (bool, error) {
	terminal := []app.RunnerJobStatus{
		app.RunnerJobStatusFinished,
		app.RunnerJobStatusFailed,
		app.RunnerJobStatusTimedOut,
		app.RunnerJobStatusNotAttempted,
		app.RunnerJobStatusCancelled,
		app.RunnerJobStatusUnknown,
	}

	deployIDs := s.db.WithContext(ctx).
		Model(&app.InstallDeploy{}).
		Select("id").
		Where(app.InstallDeploy{InstallComponentID: installComponentID})

	// Unqualified columns and a subquery, not a join: RunnerJob reads resolve to
	// runner_jobs_view_v2, so naming the table breaks at runtime.
	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where(app.RunnerJob{OwnerType: "install_deploys"}).
		Where("owner_id IN (?)", deployIDs).
		Where("status NOT IN ?", terminal).
		Count(&count)
	if res.Error != nil {
		return false, fmt.Errorf("unable to check for running deploy jobs: %w", res.Error)
	}

	return count > 0, nil
}

func (s *service) createRecoveryDeploy(
	ctx *gin.Context,
	installComponentID string,
	buildID string,
	componentID string,
	installID string,
	role string,
) (*app.InstallDeploy, error) {
	deploy := app.InstallDeploy{
		Status:             app.InstallDeployStatusQueued,
		StatusDescription:  "waiting to recover the helm release",
		ComponentBuildID:   buildID,
		InstallComponentID: installComponentID,
		Type:               app.InstallDeployTypeRecover,
		Role:               role,
	}

	if res := s.db.WithContext(ctx).Create(&deploy); res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	s.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallDeploy,
		Message:     "helm release recovery started",
		Outcome:     audit.OutcomeStarted,
		InstallID:   installID,
		ComponentID: componentID,
		SubjectID:   deploy.ID,
		SubjectType: "install_deploys",
		Attrs: map[string]string{
			"deploy.id":            deploy.ID,
			"deploy.type":          string(app.InstallDeployTypeRecover),
			"component_build.id":   buildID,
			"install_component.id": installComponentID,
		},
	})

	return &deploy, nil
}
