package components

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
)

// EnsureComponent creates a component if it doesn't exist, using the shared helpers
// for full initialization (queue creation, dependencies, install components).
func EnsureComponent(ctx context.Context, db *gorm.DB, helpers *componenthelpers.Helpers, comp *config.Component, appID string, state *sync.State) error {
	_, err := getComponent(ctx, db, comp.Name, appID)
	if err == nil {
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if component %s exists", comp.Name),
			Err:         err,
		}
	}

	newComp, err := helpers.CreateComponentWithDB(ctx, db, &componenthelpers.CreateComponentParams{
		AppID:            appID,
		Name:             comp.Name,
		VarName:          comp.VarName,
		Dependencies:     comp.Dependencies,
		Labels:           comp.Labels,
		SkipDependencies: true,
		SkipQueues:       true,
	})
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create component %s", comp.Name),
			Err:         err,
		}
	}

	if state != nil {
		if state.Result == nil {
			state.Result = &sync.Result{}
		}
		state.Result.ComponentsCreated = append(state.Result.ComponentsCreated, newComp.ID)
	}

	return nil
}

// SyncComponentParams carries the dependencies and target of a single
// component sync.
type SyncComponentParams struct {
	DB          *gorm.DB
	Helpers     *componenthelpers.Helpers
	VCSHelper   *vcshelpers.Helpers
	TFClient    terraform.Client
	Component   *config.Component
	AppID       string
	AppConfigID string
	State       *sync.State

	// DispatchBuilds reuses unchanged config connections and records changed
	// components in State.Result.ComponentsScheduled for the caller to enqueue
	// build signals for (the signal packages would import-cycle from here).
	// Leave off when a later step schedules builds (branch run's builds step).
	DispatchBuilds bool
}

// SyncComponent updates a component and creates its configuration via the shared
// builders in internal/pkg/config/build, which the per-type
// Create*ComponentConfig handlers also use.
func SyncComponent(ctx context.Context, params SyncComponentParams) error {
	db, helpers, vcsHelper, tfClient := params.DB, params.Helpers, params.VCSHelper, params.TFClient
	comp, appID, appConfigID, state := params.Component, params.AppID, params.AppConfigID, params.State

	apiComp, err := getComponent(ctx, db, comp.Name, appID)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get component %s", comp.Name),
			Err:         err,
		}
	}

	updates := app.Component{
		Name:    comp.Name,
		VarName: comp.VarName,
		Type:    app.ComponentType(comp.Type.APIType()),
		Labeled: labels.Labeled{Labels: labels.Labels(comp.Labels)},
	}

	res := db.WithContext(ctx).
		Model(apiComp).
		Select("name", "var_name", "type", "labels").
		Updates(updates)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update component %s", comp.Name),
			Err:         res.Error,
		}
	}

	depIDs := []string{}
	if len(comp.Dependencies) > 0 {
		depIDs, err = helpers.GetComponentIDsWithDB(ctx, db, appID, comp.Dependencies)
		if err != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to resolve dependencies for component %s", comp.Name),
				Err:         err,
			}
		}
	}

	vcs, err := resolveVCS(ctx, db, vcsHelper, comp, appID)
	if err != nil {
		return err
	}

	terraformVersion := ""
	if comp.TerraformModule != nil {
		terraformVersion = comp.TerraformModule.TerraformVersion
	}
	if build.NeedsTerraformVersion(comp) {
		terraformVersion, err = tfClient.GetLatestVersion()
		if err != nil {
			return sync.SyncInternalErr{
				Description: "unable to fetch latest terraform version",
				Err:         err,
			}
		}
	}

	in, err := build.ComponentConnectionInputFromConfig(
		comp,
		apiComp.ID,
		appConfigID,
		depIDs,
	)
	if err != nil {
		return sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: err.Error(),
		}
	}

	ccc, err := build.ComponentConnection(in)
	if err != nil {
		return sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: err.Error(),
		}
	}

	if err := build.AttachTypeConfig(ccc, comp, vcs, terraformVersion); err != nil {
		return sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: err.Error(),
		}
	}

	if params.DispatchBuilds {
		reusableID, err := reusableConfigID(ctx, db, apiComp.ID, ccc)
		if err != nil {
			return err
		}
		if reusableID != "" {
			state.Components = append(state.Components, sync.ComponentState{
				Name:     apiComp.Name,
				Type:     comp.Type.APIType(),
				ID:       apiComp.ID,
				ConfigID: reusableID,
				Checksum: comp.Checksum,
			})
			return nil
		}
	}

	if res := db.WithContext(ctx).Create(ccc); res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create component config for %s", comp.Name),
			Err:         res.Error,
		}
	}

	if params.DispatchBuilds {
		state.Result = appendScheduled(state.Result, sync.ComponentState{
			Name:     apiComp.Name,
			Type:     comp.Type.APIType(),
			ID:       apiComp.ID,
			ConfigID: ccc.ID,
			Checksum: comp.Checksum,
		})
	} else if comp.ExternalImage != nil {
		// Every CCC needs a build behind it. Reuse the previous CCC's Active
		// build when nothing changed; otherwise pre-create a queued build for
		// the branch run's builds step to adopt and execute via queuebuild.
		found, reusableBuildID, err := reusableActiveBuildID(ctx, db, apiComp.ID, ccc)
		if err != nil {
			return err
		}
		if found {
			if err := db.WithContext(ctx).
				Model(&app.ComponentConfigConnection{}).
				Where("id = ?", ccc.ID).
				Update("latest_build_id", reusableBuildID).Error; err != nil {
				return sync.SyncInternalErr{
					Description: fmt.Sprintf("unable to reuse build for component %s", comp.Name),
					Err:         err,
				}
			}
		} else if _, err := helpers.CreateComponentBuildInTx(ctx, db, apiComp.ID, false, nil); err != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to queue build for component %s", comp.Name),
				Err:         err,
			}
		}
	}

	state.Components = append(state.Components, sync.ComponentState{
		Name:     apiComp.Name,
		Type:     comp.Type.APIType(),
		ID:       apiComp.ID,
		ConfigID: ccc.ID,
		Checksum: comp.Checksum,
	})

	return nil
}

// reusableActiveBuildID returns the previous config connection's build ID when
// the incoming config is unchanged and that build is Active. retuurns false when
// no reusable build.
func reusableActiveBuildID(ctx context.Context, db *gorm.DB, cmpID string, incoming *app.ComponentConfigConnection) (bool, string, error) {
	if incoming.Checksum == "" || build.RequiresFreshBuild(incoming) {
		return false, "", nil
	}

	var prev app.ComponentConfigConnection
	res := db.WithContext(ctx).
		Select("id", "checksum", "latest_build_id").
		Where(app.ComponentConfigConnection{ComponentID: cmpID}).
		Where("id <> ?", incoming.ID).
		Order("created_at DESC").
		First(&prev)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to look up previous config for component %s", cmpID),
			Err:         res.Error,
		}
	}

	if prev.Checksum != incoming.Checksum || !prev.LatestBuildID.Valid {
		return false, "", nil
	}

	var bld app.ComponentBuild
	res = db.WithContext(ctx).
		Select("id", "status").
		Where(app.ComponentBuild{ID: prev.LatestBuildID.String}).
		First(&bld)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to look up previous build for component %s", cmpID),
			Err:         res.Error,
		}
	}

	if bld.Status != app.ComponentBuildStatusActive {
		return false, "", nil
	}

	return true, bld.ID, nil
}

// reusableConfigID returns the latest config connection's ID when it matches
// the incoming checksum and has a non-failed build, "" when a fresh connection
// is needed. Reusing (not skip-building a fresh one) keeps the invariant that
// every config connection has a build behind it, which CCC pinning depends on.
func reusableConfigID(ctx context.Context, db *gorm.DB, cmpID string, incoming *app.ComponentConfigConnection) (string, error) {
	checksum := incoming.Checksum
	if checksum == "" || build.RequiresFreshBuild(incoming) {
		return "", nil
	}

	var latest app.ComponentConfigConnection
	res := db.WithContext(ctx).
		Select("id", "checksum", "latest_build_id").
		Where(app.ComponentConfigConnection{ComponentID: cmpID}).
		Order("created_at DESC").
		First(&latest)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to look up latest config for component %s", cmpID),
			Err:         res.Error,
		}
	}

	if latest.Checksum != checksum || !latest.LatestBuildID.Valid {
		return "", nil
	}

	var bld app.ComponentBuild
	res = db.WithContext(ctx).
		Select("id", "status").
		Where(app.ComponentBuild{ID: latest.LatestBuildID.String}).
		First(&bld)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to look up latest build for component %s", cmpID),
			Err:         res.Error,
		}
	}

	if bld.Status == "error" {
		return "", nil
	}

	return latest.ID, nil
}

func appendScheduled(result *sync.Result, state sync.ComponentState) *sync.Result {
	if result == nil {
		result = &sync.Result{}
	}
	result.ComponentsScheduled = append(result.ComponentsScheduled, state)
	return result
}

func resolveVCS(ctx context.Context, db *gorm.DB, vcsHelper *vcshelpers.Helpers, comp *config.Component, appID string) (build.VCS, error) {
	connected, public := build.ComponentRepos(comp)
	if connected == nil && public == nil {
		return build.VCS{}, nil
	}

	var vcs build.VCS

	if connected != nil {
		var parentApp app.App
		if res := db.WithContext(ctx).
			Preload("Org").
			Preload("Org.VCSConnections").
			First(&parentApp, "id = ?", appID); res.Error != nil {
			return build.VCS{}, sync.SyncInternalErr{
				Description: "unable to get app for component vcs config",
				Err:         res.Error,
			}
		}

		cfg, err := vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      connected.Repo,
			Branch:    connected.Branch,
			Directory: connected.Directory,
		}, parentApp.Org)
		if err != nil {
			return build.VCS{}, sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create connected github vcs config for component %s", comp.Name),
				Err:         err,
			}
		}
		vcs.Github = cfg
	}

	if public != nil {
		cfg, err := vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
			Repo:      public.Repo,
			Branch:    public.Branch,
			Directory: public.Directory,
		})
		if err != nil {
			return build.VCS{}, sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create public git vcs config for component %s", comp.Name),
				Err:         err,
			}
		}
		vcs.Public = cfg
	}

	return vcs, nil
}

// EnsureComponentDependencies resolves and sets dependencies for a component.
// This must be called after all components have been created (via EnsureComponent)
// so that dependency names can be resolved to IDs.
func EnsureComponentDependencies(ctx context.Context, db *gorm.DB, helpers *componenthelpers.Helpers, comp *config.Component, appID string) error {
	if len(comp.Dependencies) == 0 {
		return nil
	}

	apiComp, err := getComponent(ctx, db, comp.Name, appID)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get component %s for dependency resolution", comp.Name),
			Err:         err,
		}
	}

	depIDs, err := helpers.GetComponentIDsWithDB(ctx, db, appID, comp.Dependencies)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to resolve dependencies for component %s", comp.Name),
			Err:         err,
		}
	}

	if err := helpers.ClearComponentDependenciesWithDB(ctx, db, apiComp.ID); err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to clear existing dependencies for component %s", comp.Name),
			Err:         err,
		}
	}

	if err := helpers.CreateComponentDependenciesWithDB(ctx, db, apiComp.ID, depIDs); err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create dependencies for component %s", comp.Name),
			Err:         err,
		}
	}

	return nil
}

// getComponent finds a component by name.
func getComponent(ctx context.Context, db *gorm.DB, name string, appID string) (*app.Component, error) {
	var comp app.Component
	res := db.WithContext(ctx).
		Where("app_id = ? AND name = ?", appID, name).
		First(&comp)

	if res.Error != nil {
		return nil, res.Error
	}

	return &comp, nil
}
