package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppBranchAppConfigs
// @Summary				    get app branch app configs
// @Description.markdown	get_app_branch_configs.md
// @Param					app_id						path	string	true	"app ID"
// @Param					app_branch_id				path	string	true	"app branch ID"
// @Param					offset						query	int		false	"offset of branches to return"	Default(0)
// @Param					limit						query	int		false	"limit of branches to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{object}	[]app.AppConfig
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/configs [get]
func (s *service) GetAppBranchAppConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")
	cfgs, err := s.getAppBranchAppConfigs(ctx, org.ID, appID, appBranchID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getAppBranchAppConfigs(ctx *gin.Context, orgID, appID, appBranchID string) ([]app.AppConfig, error) {
	configs := make([]app.AppConfig, 0)

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.AppConfig{
			OrgID:       orgID,
			AppID:       appID,
			AppBranchID: generics.NewNullString(appBranchID),
		}).
		Order("created_at desc").
		Find(&configs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branch configs: %w", res.Error)
	}

	configs, err := db.HandlePaginatedResponse(ctx, configs)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", err)
	}

	return configs, nil
}
