package branches

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func Sync(ctx context.Context, db *gorm.DB, appsHelper *appshelpers.Helpers, cfg *config.AppConfig, appID string, state *sync.State) error {
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
		if err := syncSingleBranch(ctx, db, appsHelper, branchCfg, existingByName, appID, state); err != nil {
			return err
		}
	}

	return nil
}

// Validate runs the branch checks that would otherwise only surface once the
// branches step writes, which is now last. Without it a typo in a branch block
// only fails after every component, action and runbook has synced and their
// builds have been dispatched.
//
// Runbook names are checked against runbooks that already exist plus those
// declared in this same config, because the runbook steps that create them run
// before the branches write step but after this one.
func Validate(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID string) error {
	branches := getAllBranches(cfg)
	if len(branches) == 0 {
		return nil
	}

	declaredRunbooks := make(map[string]struct{}, len(cfg.Runbooks))
	for _, rbk := range cfg.Runbooks {
		declaredRunbooks[rbk.Name] = struct{}{}
	}

	var existingRunbooks []app.Runbook
	if err := db.WithContext(ctx).Where(app.Runbook{AppID: appID}).Find(&existingRunbooks).Error; err != nil {
		return sync.SyncInternalErr{Description: "unable to list runbooks for branch validation", Err: err}
	}
	for _, rbk := range existingRunbooks {
		declaredRunbooks[rbk.Name] = struct{}{}
	}

	var nameToID map[string]string
	for _, branchCfg := range branches {
		for _, name := range branchCfg.PostDeployRunbooks {
			if _, ok := declaredRunbooks[name]; !ok {
				return sync.SyncErr{
					Resource:    "app-branches",
					Description: fmt.Sprintf("branch %q: unknown post_deploy_runbooks runbook name: %s", branchCfg.Name, name),
				}
			}
		}

		for _, group := range branchCfg.InstallGroups {
			if len(group.InstallNames) == 0 {
				continue
			}
			if nameToID == nil {
				var err error
				nameToID, err = resolveInstallNames(ctx, db, appID)
				if err != nil {
					return sync.SyncInternalErr{Description: "unable to resolve install names", Err: err}
				}
			}
			for _, name := range group.InstallNames {
				if _, ok := nameToID[name]; !ok {
					return sync.SyncErr{
						Resource:    "app-branches",
						Description: fmt.Sprintf("install group %q: unknown install name: %s", group.Name, name),
					}
				}
			}
		}
	}

	return nil
}

func syncSingleBranch(ctx context.Context, db *gorm.DB, appsHelper *appshelpers.Helpers, branchCfg *config.AppBranchConfig, existingByName map[string]*app.AppBranch, appID string, state *sync.State) error {
	existing, found := existingByName[branchCfg.Name]

	var branchID string
	if !found {
		branch, err := appsHelper.CreateAppBranchWithDB(ctx, db, appID, branchCfg.Name, app.AppBranchManagedByConfig)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return sync.SyncErr{
					Resource:    "app-branches",
					Description: fmt.Sprintf("app branch %q already exists (possibly soft-deleted)", branchCfg.Name),
				}
			}
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create app branch %q", branchCfg.Name),
				Err:         err,
			}
		}
		branchID = branch.ID
		if state != nil {
			if state.Result == nil {
				state.Result = &sync.Result{}
			}
			state.Result.AppBranchesCreated = append(state.Result.AppBranchesCreated, branch.ID)
		}
		existingByName[branch.Name] = branch
	} else {
		branchID = existing.ID
	}

	if err := db.WithContext(ctx).
		Model(&app.AppBranch{ID: branchID}).
		Update("managed_by", app.AppBranchManagedByConfig).Error; err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update managed_by for app branch %q", branchCfg.Name),
			Err:         err,
		}
	}

	if branchCfg.ConnectedRepo == nil && branchCfg.PublicRepo == nil {
		// No repo means no AppBranchConfig row is written at all, so anything
		// configured below would be silently discarded. Fail loudly rather than
		// reporting a successful sync of settings that were dropped.
		if len(branchCfg.PostDeployRunbooks) > 0 {
			return sync.SyncErr{
				Resource:    "app-branches",
				Description: fmt.Sprintf("branch %q sets post_deploy_runbooks but has no connected_repo or public_repo; post-deploy runbooks require a tracked repo", branchCfg.Name),
			}
		}
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
	if branchCfg.Preview != nil && branchCfg.Preview.InstallName != "" {
		if nameToID == nil {
			var err error
			nameToID, err = resolveInstallNames(ctx, db, appID)
			if err != nil {
				return sync.SyncInternalErr{
					Description: "unable to resolve preview install names",
					Err:         err,
				}
			}
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

	previewConfig, err := buildPreviewConfig(branchCfg, nameToID)
	if err != nil {
		return err
	}

	postDeployRunbookIDs, err := resolvePostDeployRunbooks(ctx, db, appID, branchCfg)
	if err != nil {
		return err
	}

	// Config-as-code is declarative: an omitted field in the TOML means "unset",
	// so always pass non-nil pointers rather than inheriting the previous config.
	ignoreChanges := &appshelpers.IgnoreChangesSettings{
		Regex:                &branchCfg.IgnoreChangesRegex,
		SendStatusesOnIgnore: &branchCfg.SendStatusesOnIgnore,
	}

	if _, err := appsHelper.CreateAppBranchConfigWithDB(ctx, db, branchID, connectedGithubVCSConfig, publicGitVCSConfig, installGroups, &postDeployRunbookIDs, ignoreChanges, previewConfig); err != nil {
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
			Name:                         group.Name,
			Order:                        order,
			InstallIDs:                   installIDs,
			AutoApproveOnPoliciesPassing: group.AutoApproveOnPoliciesPassing,
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

func buildPreviewConfig(branchCfg *config.AppBranchConfig, nameToID map[string]string) (*app.AppBranchPreviewConfig, error) {
	if branchCfg.Preview == nil {
		return nil, nil
	}
	p := branchCfg.Preview
	out := app.AppBranchPreviewConfig{
		Mode: app.AppBranchRunPreviewMode(p.Mode),
	}
	if out.Mode == "" {
		out.Mode = app.AppBranchRunPreviewModePlanOnly
	}
	if p.InstallID != "" {
		id := p.InstallID
		out.InstallID = &id
	}
	if p.InstallName != "" {
		id, ok := nameToID[p.InstallName]
		if !ok {
			return nil, sync.SyncErr{
				Resource:    "app-branches",
				Description: fmt.Sprintf("branch %q preview: unknown install name: %s", branchCfg.Name, p.InstallName),
			}
		}
		out.InstallID = &id
		name := p.InstallName
		out.InstallName = &name
	}
	if len(p.LabelSelector) > 0 {
		out.LabelSelector = &labels.Selector{MatchLabels: labels.Labels(p.LabelSelector)}
	}
	if p.SetStatuses != nil {
		out.SetStatuses = *p.SetStatuses
	} else {
		out.SetStatuses = true
	}
	if p.Comment != nil {
		out.Comment = *p.Comment
	} else {
		out.Comment = true
	}
	if err := out.Validate(); err != nil {
		return nil, sync.SyncErr{Resource: "app-branches", Description: err.Error()}
	}
	return &out, nil
}

// resolvePostDeployRunbooks maps the branch's runbook names to IDs. This runs
// after the runbook sync steps, so runbooks declared in this same config already
// exist alongside any pre-existing ones.
func resolvePostDeployRunbooks(ctx context.Context, db *gorm.DB, appID string, branchCfg *config.AppBranchConfig) ([]string, error) {
	if len(branchCfg.PostDeployRunbooks) == 0 {
		return nil, nil
	}

	var runbooks []app.Runbook
	if err := db.WithContext(ctx).Where(app.Runbook{AppID: appID}).Find(&runbooks).Error; err != nil {
		return nil, sync.SyncInternalErr{
			Description: "unable to list runbooks for post-deploy runbook resolution",
			Err:         err,
		}
	}

	nameToID := make(map[string]string, len(runbooks))
	for _, rbk := range runbooks {
		nameToID[rbk.Name] = rbk.ID
	}

	runbookIDs := make([]string, 0, len(branchCfg.PostDeployRunbooks))
	for _, name := range branchCfg.PostDeployRunbooks {
		id, ok := nameToID[name]
		if !ok {
			return nil, sync.SyncErr{
				Resource:    "app-branches",
				Description: fmt.Sprintf("branch %q: unknown post_deploy_runbooks runbook name: %s", branchCfg.Name, name),
			}
		}
		runbookIDs = append(runbookIDs, id)
	}

	return runbookIDs, nil
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
