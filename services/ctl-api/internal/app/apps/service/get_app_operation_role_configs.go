package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GetAppOperationRoleConfigs
// @Summary				get operation role configs
// @Description			Get all operation role configs for an app
// @Tags					apps
// @Accept					json
// @Param					app_id	path	string	true	"app ID"
// @Param					operation_role_config_id	path	string	true	"operation role config ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.AppOperationRoleConfig
// @Router					/v1/apps/{app_id}/operation-role-configs/{operation_role_config_id} [get]
func (s *service) GetAppOperationRoleConfigs(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	operationRoleConfigID := ctx.Param("operation_role_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	appOperationRoleConfig, err := s.getAppOperationRoleConfig(ctx, currentApp.ID, operationRoleConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(
			err,
			"unable to fetch app operations role config for for given app id and operations config id",
		))
	}

	ctx.JSON(http.StatusOK, appOperationRoleConfig)
}

func (s *service) getAppOperationRoleConfig(ctx *gin.Context, appID string, operationRoleConfigID string) (*app.AppOperationRoleConfig, error) {
	var appOperationRoleConfig app.AppOperationRoleConfig
	res := s.db.WithContext(ctx).
		Where(app.AppOperationRoleConfig{
			AppID: appID,
			ID:    operationRoleConfigID,
		}).
		Preload("Rules").
		First(&appOperationRoleConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find operation role configs: %w", res.Error)
	}
	return &appOperationRoleConfig, nil
}
