package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateAppSecretRequest struct {
	Name  string `json:"name" validate:"required,entity_name"`
	Value string `json:"value" validate:"required"`
}

func (c *CreateAppSecretRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateAppSecretV2
// @Summary				create an app secret
// @Description.markdown	create_app_secret.md
// @Tags					apps
// @Accept					json
// @Param					req		body	CreateAppSecretRequest	true	"Input"
// @Param					app_id	path	string					true	"app ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSecret
// @Router					/v1/apps/{app_id}/secrets [post]
func (s *service) CreateAppSecretV2(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSecretRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createAppSecret(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app secret: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app)
}

//		@ID						CreateAppSecret
//		@Summary				create an app secret
//		@Description.markdown	create_app_secret.md
//		@Tags					apps
//		@Accept					json
//		@Param					req		body	CreateAppSecretRequest	true	"Input"
//		@Param					app_id	path	string					true	"app ID"
//		@Produce				json
//		@Security				APIKey
//		@Security				OrgID
//	 @Deprecated     true
//		@Failure				400	{object}	stderr.ErrResponse
//		@Failure				401	{object}	stderr.ErrResponse
//		@Failure				403	{object}	stderr.ErrResponse
//		@Failure				404	{object}	stderr.ErrResponse
//		@Failure				500	{object}	stderr.ErrResponse
//		@Success				201	{object}	app.AppSecret
//		@Router					/v1/apps/{app_id}/secret [post]
func (s *service) CreateAppSecret(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSecretRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createAppSecret(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app secret: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app)
}

func (s *service) createAppSecret(ctx context.Context, appID string, req *CreateAppSecretRequest) (*app.AppSecret, error) {
	sec := app.AppSecret{
		AppID: appID,
		Name:  req.Name,
		Value: req.Value,
	}

	res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "app_id"}, {Name: "deleted_at"}},
			UpdateAll: true}).
		Create(&sec)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app secret: %w", res.Error)
	}

	return &sec, nil
}
