package service

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetInstallComponentsSummary
// @Summary				get an installs components summary
// @Description.markdown	get_install_components_summary.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
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

	install, err := s.getInstallByID(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	err = s.populateInstallComponentsWithDeploys(ctx, install)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to populate install components with deploys: %w", err))
		return
	}

	appID := filterAppID(install.InstallComponents)
	if appID == "" {
		ctx.Error(fmt.Errorf("unable to get app ID: %w", err))
		return
	}

	// Extract component IDs and app ID
	cmpIDs := make([]string, 0, len(install.InstallComponents))
	for _, ic := range install.InstallComponents {
		cmpIDs = append(cmpIDs, ic.ComponentID)
	}

	// Fetch the latest builds for the components
	builds, err := s.componentHelpers.GetComponentLatestBuilds(ctx, cmpIDs...)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}

	// Build visitedComps using logic from GetConfigGraph
	visitedComps := make(map[string]bool)
	for _, conn := range install.AppConfig.ComponentConfigConnections {
		visitedComps[conn.ComponentID] = true
	}

	allComps, err := s.appsHelpers.GetAppComponentsAtConfigVersion(ctx, install.AppID, install.AppConfig.Version)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app components: %w", err))
		return
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

	allCfgs := append(install.AppConfig.ComponentConfigConnections, missingComps...)

	compMap := buildInstallComponentConfig(allCfgs, install.InstallComponents)
	depComps, err := s.buildDependentComponents(compMap, allComps)
	if err != nil {
		ctx.Error(err)
		return
	}

	installSummary := s.buildSummary(install.InstallComponents, compMap, builds, depComps)

	paginatedSummary, err := db.HandlePaginatedResponse(ctx, installSummary)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to paginate install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, paginatedSummary)
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

func (s *service) getInstallByID(ctx *gin.Context, installID string) (*app.Install, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("install_components.created_at DESC")
		}).
		Preload("InstallComponents.Component").
		Preload("InstallComponents.TerraformWorkspace").
		Preload("AppConfig").
		Preload("AppConfig.ComponentConfigConnections").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}
	return install, nil
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

func (s *service) buildSummary(installComponents []app.InstallComponent, compMap map[string]*app.ComponentConfigConnection, builds []app.ComponentBuild, depComps map[string][]app.Component) []app.InstallComponentSummary {
	summaries := make([]app.InstallComponentSummary, 0, len(installComponents))
	for _, ic := range installComponents {
		deploy := filterLatestDeploy(ic.InstallDeploys)
		deployStatus := app.InstallDeployStatusUnknown
		deployStatusDescription := ""
		if deploy != nil {
			deployStatus = deploy.Status
			deployStatusDescription = deploy.StatusDescription
		}

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
			DeployStatus:            deployStatus,
			DeployStatusDescription: deployStatusDescription,
			BuildStatus:             buildStatus,
			BuildStatusDescription:  buildStatusDescription,
			ComponentConfig:         compMap[ic.ComponentID],
			Dependencies:            depComps[ic.ComponentID],
		})

	}

	return summaries
}

func (s *service) populateInstallComponentsWithDeploys(ctx *gin.Context, install *app.Install) error {
	installComponentIDs := make([]string, 0, len(install.InstallComponents))
	for _, component := range install.InstallComponents {
		installComponentIDs = append(installComponentIDs, component.ID)
	}

	latestDeploys, err := s.getLatestInstallsDeploys(ctx, installComponentIDs...)
	if err != nil {
		return fmt.Errorf("unable to get latest installs deploys: %w", err)
	}

	for i := range install.InstallComponents {
		component := &install.InstallComponents[i]
		for _, deploy := range latestDeploys {
			if component.ID == deploy.InstallComponentID {
				component.InstallDeploys = append(component.InstallDeploys, deploy)
			}
		}
	}

	return nil
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
