package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetInstallComponents
// @Summary				get an installs components
// @Description.markdown	get_install_components.md
// @Param					install_id					path	string	true	"install ID"
// @Param					types						query	string	false	"component types to filter by"
// @Param         q					query	string	false	"search query for component name or ID"
// @Param					labels						query	string	false	"label filter (key:value,key:value)"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallComponent
// @Router					/v1/installs/{install_id}/components [GET]
func (s *service) GetInstallComponents(ctx *gin.Context) {
	appID := ctx.Param("install_id")
	types := ctx.Query("types")
	q := ctx.Query("q")
	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))
	var typesSlice []string
	if types != "" {
		typesSlice = pq.StringArray(strings.Split(types, ","))
	}
	installComponents, err := s.getInstallComponents(ctx, appID, q, typesSlice, lbls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponents)
}

func (s *service) getInstallComponents(ctx *gin.Context, installID, q string, types []string, lbls labels.Labels) ([]app.InstallComponent, error) {
	paginatedComponents := []app.InstallComponent{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Scopes(labels.WithLabels("components.labels", lbls)).
		Joins("JOIN components ON components.id = install_components.component_id").
		Order("created_at DESC")

	if len(types) > 0 {
		tx = tx.
			Where("components.type IN ?", types)
	}

	if q != "" {
		tx = tx.
			Where("components.name ILIKE ? OR components.id = ?", "%"+q+"%", q)
	}

	tx = tx.Preload("Component").
		Preload("TerraformWorkspace").
		Where("install_id = ?", installID).
		Find(&paginatedComponents)

	if tx.Error != nil {
		return nil, fmt.Errorf("unable to query install components: %w", tx.Error)
	}

	if len(paginatedComponents) > 0 {
		var componentIDs []string
		for i := range paginatedComponents {
			componentIDs = append(componentIDs, paginatedComponents[i].ID)
		}

		allDriftedObjects := make([]app.DriftedObject, 0)
		res := s.db.WithContext(ctx).
			Where("install_component_id IN ?", componentIDs).
			Find(&allDriftedObjects)

		if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("unable to get drifted objects: %w", res.Error)
		}

		// Create a map of install component ID to drifted object
		driftedObjectByComponentID := make(map[string]app.DriftedObject)
		for _, obj := range allDriftedObjects {
			if obj.InstallComponentID != nil {
				// Just keep the last one if there are multiple
				driftedObjectByComponentID[*obj.InstallComponentID] = obj
			}
		}

		// Set the single drifted object for each component
		for i := range paginatedComponents {
			if obj, ok := driftedObjectByComponentID[paginatedComponents[i].ID]; ok {
				paginatedComponents[i].DriftedObject = obj
			}
		}
	}

	paginatedComponents, err := db.HandlePaginatedResponse(ctx, paginatedComponents)
	if err != nil {
		return nil, fmt.Errorf("failed to paginate install components: %w", err)
	}

	ptrs := make([]*app.InstallComponent, len(paginatedComponents))
	for i := range paginatedComponents {
		ptrs[i] = &paginatedComponents[i]
	}
	if err := s.populateComponentEnabled(ctx, installID, ptrs); err != nil {
		return nil, err
	}

	return paginatedComponents, nil
}

// populateComponentEnabled resolves and sets the Enabled field on each
// toggleable install component from the install's current input values
// (falling back to the component's default_enabled). Non-toggleable components
// are left untouched (Enabled stays nil).
func (s *service) populateComponentEnabled(ctx context.Context, installID string, comps []*app.InstallComponent) error {
	if len(comps) == 0 {
		return nil
	}

	var appConfigID string
	if err := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_config_id").
		Where("id = ?", installID).
		Scan(&appConfigID).Error; err != nil {
		return fmt.Errorf("unable to get install app config: %w", err)
	}
	if appConfigID == "" {
		return nil
	}

	var cccs []app.ComponentConfigConnection
	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where("app_config_id = ?", appConfigID).
		Find(&cccs).Error; err != nil {
		return fmt.Errorf("unable to get component config connections: %w", err)
	}
	cccByComp := make(map[string]*app.ComponentConfigConnection, len(cccs))
	for i := range cccs {
		cccByComp[cccs[i].ComponentID] = &cccs[i]
	}

	// Connections are only re-created when a component's checksum changes, so an
	// app config produced by a no-op sync has no connection rows of its own and
	// instead reuses connections pinned to an earlier config in its lineage. Mirror
	// GetFullAppConfig and resolve any missing components via the latest-configs view.
	var missingComponentIDs []string
	for _, comp := range comps {
		if _, ok := cccByComp[comp.ComponentID]; !ok {
			missingComponentIDs = append(missingComponentIDs, comp.ComponentID)
		}
	}
	if len(missingComponentIDs) > 0 {
		var fallbackCccs []app.ComponentConfigConnection
		if err := s.db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			Preload("Component").
			Where("component_id IN ?", missingComponentIDs).
			Find(&fallbackCccs).Error; err != nil {
			return fmt.Errorf("unable to get fallback component config connections: %w", err)
		}
		for i := range fallbackCccs {
			if _, ok := cccByComp[fallbackCccs[i].ComponentID]; !ok {
				cccByComp[fallbackCccs[i].ComponentID] = &fallbackCccs[i]
			}
		}
	}

	enabledInputs := map[string]*string{}
	var inputs app.InstallInputs
	if err := s.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&inputs).Error; err == nil {
		enabledInputs = inputs.Values
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("unable to get install inputs: %w", err)
	}

	resolver := app.NewComponentEnablementResolver(cccByComp, enabledInputs)
	for _, comp := range comps {
		ccc := cccByComp[comp.ComponentID]
		if ccc == nil || !ccc.IsToggleable() {
			continue
		}
		// Report effective-enabled (own toggle AND every dependency enabled) so
		// the displayed flag matches the deploy/teardown decision: a component
		// whose dependency is disabled is torn down and must read as disabled.
		enabled := resolver.EffectiveEnabled(comp.ComponentID)
		comp.Enabled = &enabled
	}

	return nil
}
