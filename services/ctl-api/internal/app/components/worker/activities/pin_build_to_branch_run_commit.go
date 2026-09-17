package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
)

// pinBuildToBranchRunCommit points a build at the commit its branch run resolved.
// Without a git ref GetBuildGitSource falls back to the tip of the component's
// branch, so a preview run builds the base branch instead of the pull request.
func (a *Activities) pinBuildToBranchRunCommit(ctx context.Context, buildID, appConfigID string) error {
	if buildID == "" || appConfigID == "" {
		return nil
	}

	var run app.AppBranchRun
	if res := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("AppBranchConfig.ConnectedGithubVCSConfig").
		Preload("AppBranchConfig.PublicGitVCSConfig").
		Where(app.AppBranchRun{AppConfigID: appConfigID}).
		Order("created_at DESC").
		First(&run); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get app branch run for app config: %w", res.Error)
	}

	if run.VCSConnectionCommit == nil {
		return nil
	}

	var build app.ComponentBuild
	if res := a.db.WithContext(ctx).
		Scopes(helpers.PreloadComponentBuildConfig).
		Where(app.ComponentBuild{ID: buildID}).
		First(&build); res.Error != nil {
		return fmt.Errorf("unable to get component build: %w", res.Error)
	}

	if !buildTracksBranchSource(&run, &build) || !buildRefCanBePinned(&build) {
		return nil
	}

	if res := a.db.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Where(app.ComponentBuild{ID: buildID}).
		Updates(map[string]any{
			"git_ref":                  run.VCSConnectionCommit.SHA,
			"vcs_connection_commit_id": run.VCSConnectionCommit.ID,
		}); res.Error != nil {
		return fmt.Errorf("unable to pin build to branch run commit: %w", res.Error)
	}

	return nil
}

// buildTracksBranchSource reports whether the build's component reads from the
// same repo and branch the app branch tracks, which is what makes the run's
// commit the right source for it. A component pointed at another repo or branch
// keeps resolving its own tip.
func buildTracksBranchSource(run *app.AppBranchRun, build *app.ComponentBuild) bool {
	connectedCfg, publicCfg := buildVCSConfigs(build)

	if branchCfg, cmpCfg := run.AppBranchConfig.ConnectedGithubVCSConfig, connectedCfg; branchCfg != nil && cmpCfg != nil {
		return strings.EqualFold(branchCfg.Repo, cmpCfg.Repo) && branchCfg.Branch == cmpCfg.Branch
	}

	if branchCfg, cmpCfg := run.AppBranchConfig.PublicGitVCSConfig, publicCfg; branchCfg != nil && cmpCfg != nil {
		return strings.EqualFold(branchCfg.Repo, cmpCfg.Repo) && branchCfg.Branch == cmpCfg.Branch
	}

	return false
}

func buildRefCanBePinned(build *app.ComponentBuild) bool {
	if build.GitRef == nil {
		return true
	}

	_, cfg := buildVCSConfigs(build)
	return cfg != nil && *build.GitRef == cfg.Branch
}

func buildVCSConfigs(build *app.ComponentBuild) (*app.ConnectedGithubVCSConfig, *app.PublicGitVCSConfig) {
	ccc := build.ComponentConfigConnection
	if ccc.ConnectedGithubVCSConfig != nil || ccc.PublicGitVCSConfig != nil {
		return ccc.ConnectedGithubVCSConfig, ccc.PublicGitVCSConfig
	}

	switch {
	case ccc.TerraformModuleComponentConfig != nil:
		return ccc.TerraformModuleComponentConfig.ConnectedGithubVCSConfig, ccc.TerraformModuleComponentConfig.PublicGitVCSConfig
	case ccc.HelmComponentConfig != nil:
		return ccc.HelmComponentConfig.ConnectedGithubVCSConfig, ccc.HelmComponentConfig.PublicGitVCSConfig
	case ccc.DockerBuildComponentConfig != nil:
		return ccc.DockerBuildComponentConfig.ConnectedGithubVCSConfig, ccc.DockerBuildComponentConfig.PublicGitVCSConfig
	case ccc.KubernetesManifestComponentConfig != nil:
		return ccc.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig, ccc.KubernetesManifestComponentConfig.PublicGitVCSConfig
	case ccc.PulumiComponentConfig != nil:
		return ccc.PulumiComponentConfig.ConnectedGithubVCSConfig, ccc.PulumiComponentConfig.PublicGitVCSConfig
	default:
		return nil, nil
	}
}
