package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateAppRunnerConfigRequest struct {
	Type    app.AppRunnerType  `json:"type"`
	EnvVars map[string]*string `json:"env_vars"`
}

func (c *CreateAppRunnerConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateAppRunnerConfig
// @Summary	create an app runner config
// @Description.markdown	create_app_runner_config.md
// @Tags			apps
// @Accept			json
// @Param			req	body	CreateAppRunnerConfigRequest	true	"Input"
// @Produce		json
// @Param			app_id	path	string				true	"app ID"
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.AppRunnerConfig
// @Router			/v1/apps/{app_id}/runner-config [post]
func (s *service) CreateAppRunnerConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppRunnerConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	runnerConfig, err := s.createAppRunnerConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, runnerConfig)
}

func (s *service) createAppRunnerConfig(ctx context.Context, appID string, req *CreateAppRunnerConfigRequest) (*app.AppRunnerConfig, error) {
	appRunnerConfig := app.AppRunnerConfig{
		AppID:   appID,
		EnvVars: pgtype.Hstore(req.EnvVars),
		Type:    req.Type,
	}
	res := s.db.WithContext(ctx).
		Create(&appRunnerConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	// update the runner configs on all installs in the app
	res = s.db.WithContext(ctx).Model(&app.Install{}).
		Where("app_id = ?", appID).
		Update("app_runner_config_id", appRunnerConfig.ID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app installs to reference new runner config: %w", res.Error)
	}

	return &appRunnerConfig, nil
}
