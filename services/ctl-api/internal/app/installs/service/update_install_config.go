package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallConfigRequest struct {
	ApprovalOption          *app.InstallApprovalOption `json:"approval_option"`
	Telemetry               *config.InstallTelemetry   `json:"telemetry,omitempty"`
	Labels                  map[string]string          `json:"labels,omitempty"`
	VPCNestedTemplateURL    *string                    `json:"vpc_nested_template_url,omitempty"`
	RunnerNestedTemplateURL *string                    `json:"runner_nested_template_url,omitempty"`
	CustomNestedStacks      []config.CustomNestedStack `json:"custom_nested_stacks,omitempty"`
	RunnerImageAutoUpdate   *bool                      `json:"runner_image_auto_update,omitempty" extensions:"x-nullable"`
	RunnerImageID           *string                    `json:"runner_image_id,omitempty"`
	ActionTimeout           *time.Duration             `json:"action_timeout,omitempty" swaggertype:"primitive,integer" extensions:"x-nullable"`

	// UseRunner changes the install's desired runner mode: nil leaves it
	// unchanged. Changing it enqueues a reprovision_stack workflow and is rejected with a 409
	// if the install has any unfinished workflow, or if the install's latest
	// stack outputs report the runner as disabled. Setting it to false also
	// requires a cloud_connection_id on the install, an AWS install, and the
	// org's runnerless-installs feature, exactly like install creation.
	UseRunner *bool `json:"use_runner,omitempty" extensions:"x-nullable"`
}

func (c *UpdateInstallConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	if err := (helpers.CreateInstallConfigParams{ActionTimeout: c.ActionTimeout}).ValidateActionTimeout(); err != nil {
		return err
	}
	return nil
}

// @ID						UpdateInstallConfig
// @Summary				update an install config
// @Description.markdown	update_install_config.md
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Param					config_id	path	string	true	"config ID"
// @Accept					json
// @Param					req	body	UpdateInstallConfigRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallConfig
// @Router					/v1/installs/{install_id}/configs/{config_id} [patch]
func (s *service) UpdateInstallConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	configID := ctx.Param("config_id")

	var req UpdateInstallConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.updateInstallConfig(ctx, installID, configID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) updateInstallConfig(ctx *gin.Context, installID, configID string, req *UpdateInstallConfigRequest) (*app.InstallConfig, error) {
	if err := helpers.ValidateStackOverrides(req.VPCNestedTemplateURL, req.RunnerNestedTemplateURL, req.CustomNestedStacks); err != nil {
		return nil, fmt.Errorf("invalid stack overrides: %w", err)
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	filter := app.InstallConfig{ID: configID, InstallID: installID, OrgID: orgID}

	// Build an explicit update map so GORM doesn't try to reflect into
	// complex nested types (CustomNestedStack contains map fields that
	// cause a panic in GORM's struct-based Updates).
	updates := map[string]interface{}{}
	if req.ApprovalOption != nil {
		updates["approval_option"] = *req.ApprovalOption
	}
	if req.Labels != nil {
		updates["labels"] = labels.Labels(req.Labels)
	}
	if req.VPCNestedTemplateURL != nil {
		updates["vpc_nested_template_url"] = *req.VPCNestedTemplateURL
	}
	if req.RunnerNestedTemplateURL != nil {
		updates["runner_nested_template_url"] = *req.RunnerNestedTemplateURL
	}
	if req.RunnerImageAutoUpdate != nil {
		updates["runner_image_auto_update"] = *req.RunnerImageAutoUpdate
	}
	if req.RunnerImageID != nil {
		updates["runner_image_id"] = *req.RunnerImageID
	}
	if req.ActionTimeout != nil {
		updates["action_timeout"] = *req.ActionTimeout
	}
	if req.CustomNestedStacks != nil {
		// Pre-serialize to JSON so the raw map-based update stores a valid
		// JSON array. Without this, GORM bypasses the serializer:json tag
		// and the pgx driver may produce a bare object instead of an array.
		b, err := json.Marshal(req.CustomNestedStacks)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal custom_nested_stacks: %w", err)
		}
		updates["custom_nested_stacks"] = string(b)
	}

	patch := cctx.PatcherFromContext(ctx)
	if req.Telemetry != nil && (req.Telemetry.Enabled != nil || (patch != nil && slices.Contains(patch.SelectFields, "telemetry.enabled"))) {
		updates["telemetry_enabled"] = req.Telemetry.Enabled
	}

	var workflowMetadata map[string]string
	if req.UseRunner != nil {
		metadata, err := s.prepareUseRunnerChange(ctx, installID, *req.UseRunner)
		if err != nil {
			return nil, err
		}
		if metadata != nil {
			updates["use_runner"] = req.UseRunner
			workflowMetadata = metadata
		}
	}

	if len(updates) > 0 {
		res := s.db.WithContext(ctx).Model(&app.InstallConfig{}).Where(filter).Updates(updates)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to patch install config: %w", res.Error)
		}
		if res.RowsAffected != 1 {
			return nil, fmt.Errorf("install config not found: %w", gorm.ErrRecordNotFound)
		}
	}
	installConfig := &app.InstallConfig{}
	if err := s.db.WithContext(ctx).Where(filter).First(installConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to reload install config: %w", err)
	}

	if workflowMetadata != nil {
		if _, err := s.createUseRunnerChangeWorkflow(ctx, installID, workflowMetadata); err != nil {
			return nil, err
		}
	}

	return installConfig, nil
}

// prepareUseRunnerChange validates a use_runner change and returns the
// workflow metadata to use for the reprovision_stack workflow it triggers, or
// nil if the request doesn't actually change the install's current desired
// value (a no-op that shouldn't enqueue anything).
func (s *service) prepareUseRunnerChange(ctx *gin.Context, installID string, desired bool) (map[string]string, error) {
	existingCfg, err := s.helpers.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install config: %w", err)
	}
	if existingCfg.UsesRunner() == desired {
		return nil, nil
	}

	var install app.Install
	if err := s.db.WithContext(ctx).Where("id = ?", installID).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	// A disabled runner freezes the install's stack in an unknown state --
	// reject a use_runner change in either direction until it's resolved via
	// reprovisioning, which is the only thing that can clear it.
	var installStack app.InstallStack
	if err := s.db.WithContext(ctx).Where(app.InstallStack{InstallID: installID}).First(&installStack).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get install stack: %w", err)
	}
	if installStack.InstallStackOutputs.RunnerDisabled() {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("use_runner cannot change while the install's runner is disabled"),
			Description: "This install's runner is disabled, so its runner mode can't be changed. Re-enable it in the install stack, then try again.",
		}
	}

	if !desired {
		cloudConnectionID := ""
		if install.CloudConnectionID != nil {
			cloudConnectionID = *install.CloudConnectionID
		}
		if err := s.helpers.ValidateUseRunnerFalse(ctx, cloudConnectionID, install.AWSAccount != nil); err != nil {
			return nil, err
		}
	}

	unfinished, err := s.helpers.HasUnfinishedInstallWorkflow(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to check for unfinished workflows: %w", err)
	}
	if unfinished {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("install has an unfinished workflow"),
			Description: "This install has a workflow still in progress. Wait for it to finish, then try again.",
		}
	}

	changeDirection := "to_runner"
	if !desired {
		changeDirection = "to_runnerless"
	}
	return map[string]string{app.WorkflowMetadataKeyUseRunnerChange: changeDirection}, nil
}

// createUseRunnerChangeWorkflow creates and enqueues a reprovision_stack
// workflow for a use_runner change, exactly like ReprovisionInstallStack.
func (s *service) createUseRunnerChangeWorkflow(ctx *gin.Context, installID string, metadata map[string]string) (string, error) {
	wf, err := s.helpers.CreateWorkflowWithRole(ctx,
		installID,
		app.WorkflowTypeReprovisionStack,
		metadata,
		false,
		"",
	)
	if err != nil {
		return "", fmt.Errorf("unable to create reprovision stack workflow: %w", err)
	}
	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		return "", fmt.Errorf("unable to get install workflows queue: %w", err)
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(wf.ID), wf.ID, "install_workflows"); err != nil {
		return "", fmt.Errorf("enqueue signal: %w", err)
	}
	return wf.ID, nil
}
