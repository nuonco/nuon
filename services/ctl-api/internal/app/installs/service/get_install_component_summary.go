package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallComponentsSummary
// @Summary				get an installs components summary
// @Description.markdown	get_install_components_summary.md
// @Param					install_id					path	string	true	"install ID"
// @Param					types						query	string	false	"component types to filter by"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param         q					query	string	false	"search query for component name"
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
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
// @Success				200	{array}		app.InstallComponentSummary
// @Router					/v1/installs/{install_id}/components/summary [GET]
func (s *service) GetInstallComponentSummary(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	types := ctx.Query("types")
	q := ctx.Query("q")
	var typesSlice []string
	if types != "" {
		typesSlice = pq.StringArray(strings.Split(types, ","))
	}

	install, err := s.getInstallByID(ctx, installID, typesSlice, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	// pagination helper relies on providing +1 results to determine if there are more results.
	// summary wasnt really designed for this. so we decrement here using the HandlePaginatedResponse.
	ics, err := db.HandlePaginatedResponse(ctx, install.InstallComponents)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to paginate install components: %w", err))
		return
	}

	if len(ics) < 1 {
		ctx.JSON(http.StatusOK, []app.InstallComponentSummary{})
		return
	}

	appID := filterAppID(ics)
	if appID == "" {
		ctx.Error(fmt.Errorf("unable to get app ID from install components"))
		return
	}

	cmpIDs := make([]string, 0, len(ics))
	for _, ic := range install.InstallComponents {
		cmpIDs = append(cmpIDs, ic.ComponentID)
	}

	// Fetch the latest builds for the components
	builds, err := s.componentHelpers.GetComponentLatestBuilds(ctx, cmpIDs...)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}

	allComps, err := s.appsHelpers.GetAppComponentsAtConfigVersion(ctx, install.AppID, install.AppConfig.Version)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app components: %w", err))
		return
	}

	missingComps := s.getMissingComponents(install.AppConfig.ComponentConfigConnections, allComps)
	allCfgs := append(install.AppConfig.ComponentConfigConnections, missingComps...)

	fillComponentType(allComps, allCfgs)

	compMap := buildInstallComponentConfig(allCfgs, ics)
	depComps, err := s.buildDependentComponents(compMap, allComps)
	if err != nil {
		ctx.Error(err)
		return
	}

	installSummary := s.buildSummary(ctx, ics, compMap, builds, depComps)

	ctx.JSON(http.StatusOK, installSummary)
}

func buildInstallComponentConfig(installComponentConfigurations []app.ComponentConfigConnection, installComponents []app.InstallComponent) map[string]*app.ComponentConfigConnection {
	compMap := make(map[string]*app.ComponentConfigConnection)
	for _, installComponent := range installComponents {
		for _, installComponentConfig := range installComponentConfigurations {
			if installComponentConfig.ComponentID == installComponent.ComponentID {
				compMap[installComponent.ComponentID] = &installComponentConfig
			}
		}
	}
	return compMap
}

func (s *service) getInstallByID(ctx *gin.Context, installID string, types []string, q string) (*app.Install, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AppConfig").
		Preload("AppConfig.ComponentConfigConnections").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	iCmps, err := s.getInstallsComponentsInBatches(ctx, installID, types, q)
	if err != nil {
		return nil, fmt.Errorf("unable to get install components in batches: %w", err)
	}

	install.InstallComponents = iCmps

	return install, nil
}

func (s *service) getInstallsComponentsInBatches(ctx *gin.Context, installID string, types []string, q string) ([]app.InstallComponent, error) {
	installComponents := make([]app.InstallComponent, 0)
	batchSize := 10
	offset := 0
	hasMore := true

	for hasMore {
		var installComponentsBatch []app.InstallComponent
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
			Limit(batchSize).
			Offset(offset).
			Find(&installComponents)

		if tx.Error != nil {
			return nil, fmt.Errorf("unable to get install components for install %s: %w", installID, tx.Error)
		}

		installComponents = append(installComponents, installComponentsBatch...)

		if len(installComponentsBatch) < batchSize {
			hasMore = false
		} else {
			offset += batchSize
		}
	}

	return installComponents, nil
}

func filterLatestDeploy(deploys []app.InstallDeploy) *app.InstallDeploy {
	sort.SliceStable(deploys, func(i, j int) bool {
		return deploys[i].CreatedAt.After(deploys[j].CreatedAt)
	})
	if len(deploys) > 0 {
		return &deploys[0]
	}
	return nil
}

func filterBuildByComponentID(builds []app.ComponentBuild, componentID string) *app.ComponentBuild {
	for _, b := range builds {
		if b.ComponentID == componentID {
			return &b
		}
	}
	return nil
}

func filterAppID(installComponents []app.InstallComponent) string {
	for _, ic := range installComponents {
		if ic.Component.AppID != "" {
			return ic.Component.AppID
		}
	}
	return ""
}

func (s *service) buildSummary(ctx context.Context, installComponents []app.InstallComponent, compMap map[string]*app.ComponentConfigConnection, builds []app.ComponentBuild, depComps map[string][]app.Component) []app.InstallComponentSummary {
	summaries := make([]app.InstallComponentSummary, 0, len(installComponents))
	for _, ic := range installComponents {
		build := filterBuildByComponentID(builds, ic.ComponentID)
		buildStatus := app.ComponentBuildStatusBuilding
		buildStatusDescription := ""
		if build != nil {
			buildStatus = build.Status
			buildStatusDescription = build.StatusDescription
		}

		summaries = append(summaries, app.InstallComponentSummary{
			ID:                      ic.ID,
			ComponentID:             ic.ComponentID,
			ComponentName:           ic.Component.Name,
			DeployStatus:            app.InstallDeployStatus(ic.Status),
			DeployStatusDescription: ic.StatusDescription,
			BuildStatus:             buildStatus,
			BuildStatusDescription:  buildStatusDescription,
			ComponentConfig:         compMap[ic.ComponentID],
			Dependencies:            depComps[ic.ComponentID],
		})

	}

	return summaries
}

func (s *service) buildDependentComponents(compMap map[string]*app.ComponentConfigConnection, comps []app.Component) (map[string][]app.Component, error) {
	depComps := make(map[string][]app.Component)

	compLookup := make(map[string]app.Component)
	for _, comp := range comps {
		compLookup[comp.ID] = comp
	}

	for compID, config := range compMap {
		if config == nil {
			continue
		}

		for _, depID := range pq.StringArray(config.ComponentDependencyIDs) {
			if depComp, exists := compLookup[depID]; exists {
				depComps[compID] = append(depComps[compID], depComp)
			}
		}
	}

	return depComps, nil
}

func (s *service) getMissingComponents(existingConnections []app.ComponentConfigConnection, allComps []app.Component) []app.ComponentConfigConnection {
	visitedComps := make(map[string]bool)
	for _, conn := range existingConnections {
		visitedComps[conn.ComponentID] = true
	}

	missingComps := make([]app.ComponentConfigConnection, 0)
	for _, comp := range allComps {
		if _, exists := visitedComps[comp.ID]; !exists {
			if len(comp.ComponentConfigs) < 1 {
				continue
			}
			missingComps = append(missingComps, comp.ComponentConfigs[0])
		}
	}
	return missingComps
}

func fillComponentType(allComps []app.Component, allCfgs []app.ComponentConfigConnection) {
	compTypeMap := make(map[string]app.ComponentType)
	for _, comp := range allComps {
		compTypeMap[comp.ID] = comp.Type
	}

	for i := range allCfgs {
		if compType, exists := compTypeMap[allCfgs[i].ComponentID]; exists {
			allCfgs[i].Type = compType
		}
	}
}
