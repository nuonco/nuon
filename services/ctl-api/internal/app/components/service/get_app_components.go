package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppComponents
// @Summary				get all components for an app
// @Description.markdown	get_app_components.md
// @Param					app_id						path	string	true	"app ID"
// @Param         q                 query	string	false	"search query to filter components by name"
// @Param         types					    query	string	false	"comma-separated list of component types to filter by (e.g., terraform_module, helm_chart)"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Component
// @Router					/v1/apps/{app_id}/components [GET]
func (s *service) GetAppComponents(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	q := ctx.Query("q")
	types := ctx.Query("types")
	var typesSlice []string
	if types != "" {
		typesSlice = pq.StringArray(strings.Split(types, ","))
	}

	components, err := s.getAppComponents(ctx, appID, q, typesSlice)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app components: %w", err))
		return
	}

	//reorderedCmp, err := s.appsHelpers.OrderComponentsByDep(ctx, component)
	//if err != nil {
	//ctx.Error(fmt.Errorf("unable to order components by dependency: %w", err))
	//return
	//}

	ctx.JSON(http.StatusOK, components)
}

func (s *service) getAppComponents(ctx *gin.Context, appID, q string, types []string) ([]app.Component, error) {
	currentApp := &app.App{}

	appCfg, err := s.appsHelpers.GetAppLatestConfig(ctx, appID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest app config")
	}

	tx := s.db.WithContext(ctx).
		Preload("Components", func(db *gorm.DB) *gorm.DB {
			db = db.Scopes(scopes.WithOffsetPagination)
			db = db.Where("id in ?", appCfg.ComponentIDs)
			if q != "" {
				db = db.Where("components.name ILIKE ?", "%"+q+"%")
			}

			if len(types) > 0 {
				db = db.Where("components.type IN ?", types)
			}
			return db
		}).
		Preload("Components.ComponentConfigs").
		Preload("Components.Dependencies")

	res := tx.First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	cmps, err := db.HandlePaginatedResponse(ctx, currentApp.Components)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	currentApp.Components = cmps

	return currentApp.Components, nil
}
