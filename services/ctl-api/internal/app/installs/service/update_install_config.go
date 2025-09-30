package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallConfigRequest struct {
	ApprovalOption *app.InstallApprovalOption `json:"approval_option"`
}

func (c *UpdateInstallConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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
	installConfig := &app.InstallConfig{
		ID: configID,
	}

	res := s.db.WithContext(ctx).
		Model(&installConfig).
		Updates(req)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to patch install config: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("install config not found: %w", gorm.ErrRecordNotFound)
	}
	return installConfig, nil
}
