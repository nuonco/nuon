package service

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppComponents
// @Summary				get all components for an app
// @Description.markdown	get_app_components.md
// @Param					app_id						path	string	true	"app ID"
// @Param         q                 query	string	false	"search query to filter components by name or ID"
// @Param         types					    query	string	false	"comma-separated list of component types to filter by (e.g., terraform_module, helm_chart)"
// @Param 				component_ids		query	string	false	"comma-separated list of component IDs to filter by"
// @Param					branch_id					query	string	false	"only return the components defined by this branch's config"
// @Param					labels						query	string	false	"label filter (key:value,key:value)"
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
func (s *service) GetAppComponents(ctx *gin.Context) (any, error) {
	appID := ctx.Param("app_id")
	q := ctx.Query("q")
	types := ctx.Query("types")
	var typesSlice []string
	if types != "" {
		typesSlice = pq.StringArray(strings.Split(types, ","))
	}
	componentIDs := ctx.Query("component_ids")
	var componentIDsSlice []string
	if componentIDs != "" {
		componentIDsSlice = pq.StringArray(strings.Split(strings.TrimSpace(componentIDs), ","))
		for i, id := range componentIDsSlice {
			componentIDsSlice[i] = strings.TrimSpace(id)
		}
	}

	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))
	branchID := ctx.Query("branch_id")

	return s.getAppComponents(ctx, appID, branchID, q, typesSlice, componentIDsSlice, lbls)
}

func (s *service) getAppComponents(ctx *gin.Context, appID, branchID, q string, types []string, componentIDs []string, lbls labels.Labels) ([]app.Component, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}
	appCfg, err := s.scopedAppConfig(ctx, appID, branchID)
	if err != nil {
		if errors.Is(err, appshelpers.ErrAppBranchNotFound) {
			return nil, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []app.Component{}, nil
		}
		return nil, errors.Wrap(err, "unable to get latest app config")
	}

	var components []app.Component
	tx := query.WithContext(ctx.Request.Context()).
		Scopes(scopes.WithOffsetPagination).
		Scopes(labels.WithLabels("components.labels", lbls)).
		Where(&app.Component{AppID: appID}).
		Where("components.id IN ?", []string(appCfg.ComponentIDs)).
		Order("components.created_at DESC, components.id DESC").
		Preload("Dependencies", &app.Component{AppID: appID, OrgID: access.OrgID()}).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Where(&app.ComponentConfigConnection{OrgID: access.OrgID()}).Scopes(scopes.WithOverrideTable(app.LatestComponentConfigConnectionsViewName))
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Where(&app.ComponentBuild{OrgID: access.OrgID()}).
				Where("component_builds.component_id IS NULL OR component_builds.component_id = '' OR component_builds.component_id = (SELECT component_id FROM component_config_connections WHERE id = component_builds.component_config_connection_id)").
				Order("component_builds.created_at DESC, component_builds.id DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit")
	if q != "" {
		tx = tx.Where("components.name ILIKE ? OR components.id = ?", "%"+q+"%", q)
	}

	if len(types) > 0 {
		tx = tx.Where("components.type IN ?", types)
	}

	if len(componentIDs) > 0 {
		tx = tx.Where("components.id IN ?", componentIDs)
	}

	res := tx.Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	cmps, err := db.HandlePaginatedResponse(ctx, components)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return cmps, nil
}

func (s *service) scopedAppConfig(ctx *gin.Context, appID, branchID string) (*app.AppConfig, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	return s.appsHelpers.GetComponentAppConfigInTx(ctx.Request.Context(), access.DB(), appID, branchID)
}
