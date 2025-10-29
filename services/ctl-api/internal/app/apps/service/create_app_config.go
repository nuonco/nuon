package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateAppConfigRequest struct {
	// not required Readme
	Readme     string `json:"readme,omitempty"`
	CLIVersion string `json:"cli_version,omitempty"`
}

func (c *CreateAppConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

// @ID						CreateAppConfigV2
// @Description.markdown	create_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/configs [post]
func (s *service) CreateAppConfigV2(ctx *gin.Context) {
	s.CreateAppConfig(ctx)
}

// @ID						CreateAppConfig
// @Description.markdown	create_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/config [post]
func (s *service) CreateAppConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createAppConfig(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app inputs config: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppConfig(ctx context.Context, orgID, appID string, req *CreateAppConfigRequest) (*app.AppConfig, error) {
	inputs := app.AppConfig{
		OrgID:             orgID,
		AppID:             appID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "sync pending",
		Readme:            req.Readme,
		CLIVersion:        req.CLIVersion,
	}

	res := s.db.WithContext(ctx).Create(&inputs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app inputs: %w", res.Error)
	}

	return &inputs, nil
}
