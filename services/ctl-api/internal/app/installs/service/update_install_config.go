package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallConfigRequest struct {
	ApprovalOption          *app.InstallApprovalOption `json:"approval_option"`
	VPCNestedTemplateURL    *string                    `json:"vpc_nested_template_url,omitempty"`
	RunnerNestedTemplateURL *string                    `json:"runner_nested_template_url,omitempty"`
	CustomNestedStacks      []config.CustomNestedStack `json:"custom_nested_stacks,omitempty"`
}

func (c *UpdateInstallConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
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

	// Build an explicit update map so GORM doesn't try to reflect into
	// complex nested types (CustomNestedStack contains map fields that
	// cause a panic in GORM's struct-based Updates).
	updates := map[string]interface{}{}
	if req.ApprovalOption != nil {
		updates["approval_option"] = *req.ApprovalOption
	}
	if req.VPCNestedTemplateURL != nil {
		updates["vpc_nested_template_url"] = *req.VPCNestedTemplateURL
	}
	if req.RunnerNestedTemplateURL != nil {
		updates["runner_nested_template_url"] = *req.RunnerNestedTemplateURL
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

	if len(updates) == 0 {
		// Nothing to update — just return the current record.
		installConfig := &app.InstallConfig{}
		if err := s.db.WithContext(ctx).First(installConfig, "id = ?", configID).Error; err != nil {
			return nil, fmt.Errorf("install config not found: %w", err)
		}
		return installConfig, nil
	}

	installConfig := &app.InstallConfig{
		ID: configID,
	}

	res := s.db.WithContext(ctx).
		Model(&installConfig).
		Updates(updates)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to patch install config: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("install config not found: %w", gorm.ErrRecordNotFound)
	}

	// Reload the full record so the response includes all fields.
	if err := s.db.WithContext(ctx).First(installConfig, "id = ?", configID).Error; err != nil {
		return nil, fmt.Errorf("unable to reload install config: %w", err)
	}
	return installConfig, nil
}
