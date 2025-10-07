package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetInstallComponents
// @Summary				get an installs components
// @Description.markdown	get_install_components.md
// @Param					install_id					path	string	true	"install ID"
// @Param					types						query	string	false	"component types to filter by"
// @Param         q					query	string	false	"search query for component name"
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
	var typesSlice []string
	if types != "" {
		typesSlice = pq.StringArray(strings.Split(types, ","))
	}
	installComponents, err := s.getInstallComponents(ctx, appID, q, typesSlice)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponents)
}

func (s *service) getInstallComponents(ctx *gin.Context, installID, q string, types []string) ([]app.InstallComponent, error) {
	paginatedComponents := []app.InstallComponent{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN components ON components.id = install_components.component_id").
		Order("created_at DESC")

	if len(types) > 0 {
		tx = tx.
			Where("components.type IN ?", types)
	}

	if q != "" {
		tx = tx.
			Where("components.name ILIKE ?", "%"+q+"%")
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

	return paginatedComponents, nil
}
