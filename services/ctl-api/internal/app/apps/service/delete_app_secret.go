package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.deleteAppSecret(ctx, orgID, ctx.Param("app_id"), ctx.Param("secret_id")); err != nil {
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
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.deleteAppSecret(ctx, orgID, ctx.Param("app_id"), ctx.Param("secret_id")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

// deleteAppSecret confines the delete to the app and org named in the URL,
// which is what the request was authorized against. Without appID a caller
// authorized on one app could delete any secret by id, in any org.
func (s *service) deleteAppSecret(ctx context.Context, orgID, appID, secretID string) error {
	res := s.db.WithContext(ctx).
		Where(app.AppSecret{
			ID:    secretID,
			AppID: appID,
			OrgID: orgID,
		}).
		Delete(&app.AppSecret{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app secret: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return stderr.ErrNotFound{
			Err:         fmt.Errorf("app secret %q not found for app %q", secretID, appID),
			Description: "app secret not found",
		}
	}

	return nil
}
