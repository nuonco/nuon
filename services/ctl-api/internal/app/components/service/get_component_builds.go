package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetComponentBuilds
// @Summary				get builds for components
// @Description.markdown	get_component_builds.md
// @Param					component_id				query	string	false	"component id to filter by"
// @Param					app_id						query	string	false	"app id to filter by"
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
// @Success				200	{array}		app.ComponentBuild
// @Router					/v1/builds [GET]
func (s *service) GetComponentBuilds(ctx *gin.Context) {
	cmpID := ctx.Query("component_id")
	appID := ctx.Query("app_id")
	if cmpID == "" && appID == "" {
		ctx.Error(fmt.Errorf("component id or app id must be passed in"))
		return
	}

	limitStr := ctx.DefaultQuery("limit", "50")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	var blds []app.ComponentBuild
	if cmpID != "" {
		blds, err = s.getComponentBuilds(ctx, cmpID)
	} else {
		blds, err = s.getAppBuilds(ctx, appID, limitVal)
	}

	if err != nil {
		ctx.Error(fmt.Errorf("unable to get builds: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, blds)
}

func (s *service) getAppBuilds(ctx *gin.Context, appID string, limit int) ([]app.ComponentBuild, error) {
	blds := []app.ComponentBuild{}

	// query all builds that belong to the component id, starting at the component to ensure the component exists
	// via the double join.
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("ComponentConfigConnection").
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection.Component").
		Joins("JOIN component_config_connections ON component_config_connections.id=component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id=component_config_connections.component_id").
		Where("components.app_id = ?", appID).
		Limit(limit).
		Order("created_at DESC").
		Find(&blds)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app builds: %w", res.Error)
	}

	blds, err := db.HandlePaginatedResponse(ctx, blds)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return blds, nil
}

func (s *service) getComponentBuilds(ctx *gin.Context, cmpID string) ([]app.ComponentBuild, error) {
	cmp := app.Component{}

	// query all builds that belong to the component id, starting at the component to ensure the component exists
	// via the double join.
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("component_builds.created_at DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit").
		Preload("ComponentConfigs.ComponentBuilds.ComponentConfigConnection").
		Preload("ComponentConfigs.ComponentBuilds.ComponentConfigConnection.Component").
		Preload("ComponentConfigs.ComponentBuilds.CreatedBy").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	blds := make([]app.ComponentBuild, 0)
	for _, cfg := range cmp.ComponentConfigs {
		blds = append(blds, cfg.ComponentBuilds...)
	}

	blds, err := db.HandlePaginatedResponse(ctx, blds)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return blds, nil
}
