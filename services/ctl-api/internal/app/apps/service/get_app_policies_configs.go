package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppPoliciesConfigs
// @Summary				get app policies configs
// @Description.markdown	get_app_policies_configs.md
// @Param					app_id	path	string	true	"app ID"
// @Param					offset	query	int		false	"offset of results to return"	Default(0)
// @Param					limit	query	int		false	"limit of results to return"	Default(10)
// @Param					page	query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{object}	[]app.AppPoliciesConfig
// @Router					/v1/apps/{app_id}/policies-configs [get]
func (s *service) GetAppPoliciesConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	configs, err := s.findAppPoliciesConfigs(ctx, org.ID, currentApp.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app policies configs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}

func (s *service) findAppPoliciesConfigs(ctx *gin.Context, orgID, appID string) ([]app.AppPoliciesConfig, error) {
	var configs []app.AppPoliciesConfig

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where("org_id = ?", orgID).
		Where("app_id = ?", appID).
		Preload("Policies").
		Order("created_at DESC").
		Find(&configs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app policies configs: %w", res.Error)
	}

	configs, err := db.HandlePaginatedResponse(ctx, configs)
	if err != nil {
		return nil, fmt.Errorf("unable to get app policies configs: %w", err)
	}

	return configs, nil
}
