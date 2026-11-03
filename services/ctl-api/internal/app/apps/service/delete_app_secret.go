package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteAppSecretV2
// @Summary				delete an app secret
// @Description.markdown	delete_app_secret.md
// @Param					app_id		path	string	true	"app ID"
// @Param					secret_id	path	string	true	"secret ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id}/secrets/{secret_id} [DELETE]
func (s *service) DeleteAppSecretV2(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.deleteAppSecret(ctx, org.ID, ctx.Param("app_id"), ctx.Param("secret_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

// @ID						DeleteAppSecret
// @Summary				delete an app secret
// @Description.markdown	delete_app_secret.md
// @Param					app_id		path	string	true	"app ID"
// @Param					secret_id	path	string	true	"secret ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id}/secret/{secret_id} [DELETE]
func (s *service) DeleteAppSecret(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.deleteAppSecret(ctx, org.ID, ctx.Param("app_id"), ctx.Param("secret_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) deleteAppSecret(ctx context.Context, orgID, appRef, secretID string) error {
	currentApp, err := s.findAppRef(ctx, orgID, appRef)
	if err != nil {
		return err
	}

	res := s.db.WithContext(ctx).
		Where(app.AppSecret{ID: secretID, OrgID: orgID, AppID: currentApp.ID}).
		Delete(&app.AppSecret{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app secret: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("app secret %q: %w", secretID, gorm.ErrRecordNotFound)
	}

	return nil
}
