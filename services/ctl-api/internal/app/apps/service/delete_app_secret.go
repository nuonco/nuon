package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						DeleteAppSecret
//	@Summary				delete an app secret
//	@Description.markdown	delete_app_secret.md
//	@Param					app_id		path	string	true	"app ID"
//	@Param					secret_id	path	string	true	"secret ID"
//	@Tags					apps
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{boolean}	true
//	@Router					/v1/apps/{app_id}/secret/{secret_id} [DELETE]
func (s *service) DeleteAppSecret(ctx *gin.Context) {
	secretID := ctx.Param("secret_id")

	err := s.deleteAppSecret(ctx, secretID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteAppSecret(ctx context.Context, secretID string) error {
	res := s.db.WithContext(ctx).
		Delete(&app.AppSecret{
			ID: secretID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app secret: %w", res.Error)
	}

	return nil
}
