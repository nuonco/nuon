package branches

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func Sync(ctx context.Context, db *gorm.DB, appsHelper *appshelpers.Helpers, cfg *config.AppConfig, appID string) error {
	branches := getAllBranches(cfg)
	if len(branches) == 0 {
		return nil
	}

	var existing []app.AppBranch
	if err := db.WithContext(ctx).Where(app.AppBranch{AppID: appID}).Find(&existing).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to list app branches",
			Err:         err,
		}
	}

	existingByName := make(map[string]*app.AppBranch, len(existing))
	for i := range existing {
		existingByName[existing[i].Name] = &existing[i]
	}

	for _, branchCfg := range branches {
		if err := syncSingleBranch(ctx, db, appsHelper, branchCfg, existingByName, appID); err != nil {
			return err
		}
	}

	return nil
}

func syncSingleBranch(ctx context.Context, db *gorm.DB, appsHelper *appshelpers.Helpers, branchCfg *config.AppBranchConfig, existingByName map[string]*app.AppBranch, appID string) error {
	existing, found := existingByName[branchCfg.Name]

	var branchID string
	if !found {
		branch, err := appsHelper.CreateAppBranch(ctx, appID, branchCfg.Name)
		if err != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create app branch %q", branchCfg.Name),
				Err:         err,
			}
		}
		branchID = branch.ID
		existingByName[branch.Name] = branch
	} else {
		branchID = existing.ID
	}

	if err := db.WithContext(ctx).
		Model(&app.AppBranch{ID: branchID}).
		Select("managed_by").
		Updates(app.AppBranch{ManagedBy: app.AppBranchManagedByConfig}).Error; err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update managed_by for app branch %q", branchCfg.Name),
			Err:         err,
		}
	}

	if branchCfg.ConnectedRepo == nil && branchCfg.PublicRepo == nil {
		return nil
	}

	var nameToID map[string]string
	for _, group := range branchCfg.InstallGroups {
		if len(group.InstallNames) > 0 {
			var err error
			nameToID, err = resolveInstallNames(ctx, db, appID)
			if err != nil {
				return sync.SyncInternalErr{
					Description: "unable to resolve install names",
					Err:         err,
				}
			}
			break
		}
	}

	var parentApp app.App
	if err := db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", appID).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to get app for VCS config",
			Err:         err,
		}
	}

	vcsHelper := appsHelper.VCSHelpers()

	var connectedGithubVCSConfig *app.ConnectedGithubVCSConfig
	if branchCfg.ConnectedRepo != nil {
		cfg, err := vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      branchCfg.ConnectedRepo.Repo,
			Branch:    branchCfg.ConnectedRepo.Branch,
			Directory: branchCfg.ConnectedRepo.Directory,
		}, parentApp.Org)
		if err != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to build connected VCS config for branch %q", branchCfg.Name),
				Err:         err,
			}
		}
		connectedGithubVCSConfig = cfg
	}

	var publicGitVCSConfig *app.PublicGitVCSConfig
	if branchCfg.PublicRepo != nil {
		cfg, err := vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
			Repo:      branchCfg.PublicRepo.Repo,
			Branch:    branchCfg.PublicRepo.Branch,
			Directory: branchCfg.PublicRepo.Directory,
		})
		if err != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to build public VCS config for branch %q", branchCfg.Name),
				Err:         err,
			}
		}
		publicGitVCSConfig = cfg
	}

	installGroups, err := buildInstallGroups(branchCfg, nameToID)
	if err != nil {
		return err
	}

	if _, err := appsHelper.CreateAppBranchConfig(ctx, branchID, connectedGithubVCSConfig, publicGitVCSConfig, installGroups); err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create config for branch %q", branchCfg.Name),
			Err:         err,
		}
	}

	return nil
}

func buildInstallGroups(branchCfg *config.AppBranchConfig, nameToID map[string]string) ([]app.AppBranchInstallGroup, error) {
	var installGroups []app.AppBranchInstallGroup
	for i, group := range branchCfg.InstallGroups {
		order := group.Order
		if order == 0 {
			order = i
		}

		installIDs := group.InstallIDs
		if len(group.InstallNames) > 0 {
			seen := make(map[string]bool, len(installIDs))
			for _, id := range installIDs {
				seen[id] = true
			}
			for _, name := range group.InstallNames {
				id, ok := nameToID[name]
				if !ok {
					return nil, sync.SyncErr{
						Resource:    "app-branches",
						Description: fmt.Sprintf("install group %q: unknown install name: %s", group.Name, name),
					}
				}
				if !seen[id] {
					installIDs = append(installIDs, id)
					seen[id] = true
				}
			}
		}

		ig := app.AppBranchInstallGroup{
			Name:           group.Name,
			Order:          order,
			InstallIDs:     installIDs,
			UseForPreviews: group.UseForPreviews,
		}

		if len(group.LabelSelector) > 0 {
			ig.LabelSelector = &labels.Selector{
				MatchLabels: labels.Labels(group.LabelSelector),
			}
		}

		installGroups = append(installGroups, ig)
	}
	return installGroups, nil
}

func resolveInstallNames(ctx context.Context, db *gorm.DB, appID string) (map[string]string, error) {
	var installs []app.Install
	if err := db.WithContext(ctx).Where(app.Install{AppID: appID}).Find(&installs).Error; err != nil {
		return nil, err
	}
	nameToID := make(map[string]string, len(installs))
	for _, inst := range installs {
		if inst.Name != "" {
			nameToID[inst.Name] = inst.ID
		}
	}
	return nameToID, nil
}

func getAllBranches(cfg *config.AppConfig) []*config.AppBranchConfig {
	var branches []*config.AppBranchConfig
	if cfg.Branch != nil {
		branches = append(branches, cfg.Branch)
	}
	branches = append(branches, cfg.Branches...)
	return branches
}
