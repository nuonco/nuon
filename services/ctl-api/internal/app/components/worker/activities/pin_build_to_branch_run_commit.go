package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func (a *Activities) pinBuildToBranchRunCommit(ctx context.Context, buildID, appConfigID string) error {
	if buildID == "" || appConfigID == "" {
		return nil
	}

	var build app.ComponentBuild
	if res := a.db.WithContext(ctx).
		Scopes(helpers.PreloadComponentBuildConfig).
		Where(app.ComponentBuild{ID: buildID}).
		First(&build); res.Error != nil {
		return fmt.Errorf("unable to get component build: %w", res.Error)
	}

	var run app.AppBranchRun
	runQuery := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("AppBranchConfig.ConnectedGithubVCSConfig").
		Preload("AppBranchConfig.PublicGitVCSConfig")
	if build.AppBranchRunID != nil {
		runQuery = runQuery.Where(app.AppBranchRun{ID: *build.AppBranchRunID})
	} else {
		runQuery = runQuery.
			Where(app.AppBranchRun{AppConfigID: appConfigID}).
			Order("created_at DESC")
	}
	res := runQuery.First(&run)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get app branch run for app config: %w", res.Error)
	}

	if buildTracksBranchSource(&run, &build) {
		if run.VCSConnectionCommit == nil || !buildRefCanBePinned(&build) {
			return nil
		}
		return a.setBuildSource(ctx, buildID, run.VCSConnectionCommit.SHA, run.VCSConnectionCommit.ID)
	}

	return a.resolveBuildOwnCommit(ctx, &build)
}

func (a *Activities) resolveBuildOwnCommit(ctx context.Context, build *app.ComponentBuild) error {
	connectedCfg, publicCfg := buildVCSConfigs(build)
	if connectedCfg == nil && publicCfg == nil {
		return nil
	}

	if build.VCSConnectionCommitID != nil {
		return nil
	}

	repo, configBranch := buildRepoRef(connectedCfg, publicCfg)
	ref := configBranch
	if build.GitRef != nil && *build.GitRef != "" {
		ref = *build.GitRef
	}
	if ref == "" {
		return nil
	}

	var (
		vcsCommit *app.VCSConnectionCommit
		fetchErr  error
	)
	if connectedCfg != nil {
		var commit *github.RepositoryCommit
		commit, fetchErr = a.vcsHelpers.GetConnectedGithubVCSConfigCommit(ctx, connectedCfg, ref)
		if fetchErr == nil {
			vcsCommit = a.vcsHelpers.GithubCommitToVCSConnectionCommit(
				commit,
				connectedCfg.ID,
				plugins.TableName(a.db, &app.ConnectedGithubVCSConfig{}),
				connectedCfg.VCSConnectionID,
			)
		}
	} else {
		var commit *github.RepositoryCommit
		commit, fetchErr = a.vcsHelpers.GetPublicGitVCSConfigCommit(ctx, publicCfg, ref)
		if fetchErr == nil {
			vcsCommit = a.vcsHelpers.GithubCommitToVCSConnectionCommit(
				commit,
				publicCfg.ID,
				plugins.TableName(a.db, &app.PublicGitVCSConfig{}),
				"",
			)
		}
	}
	if fetchErr != nil {
		return a.failBuildSource(ctx, build.ID, repo, ref, fetchErr)
	}
	if vcsCommit == nil {
		return fmt.Errorf("invalid commit data from GitHub for %s@%s", repo, ref)
	}

	return a.persistBuildSourceCommit(ctx, build.ID, vcsCommit)
}

func (a *Activities) persistBuildSourceCommit(ctx context.Context, buildID string, commit *app.VCSConnectionCommit) error {
	if err := a.db.WithContext(ctx).Create(commit).Error; err != nil {
		return fmt.Errorf("unable to create vcs commit: %w", err)
	}
	return a.setBuildSource(ctx, buildID, commit.SHA, commit.ID)
}

func (a *Activities) setBuildSource(ctx context.Context, buildID, sha, commitID string) error {
	if res := a.db.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Where(app.ComponentBuild{ID: buildID}).
		Updates(map[string]any{
			"git_ref":                  sha,
			"vcs_connection_commit_id": commitID,
			"composite_error":          nil,
		}); res.Error != nil {
		return fmt.Errorf("unable to set build source commit: %w", res.Error)
	}
	return nil
}

func (a *Activities) failBuildSource(ctx context.Context, buildID, repo, ref string, cause error) error {
	if !vcserrors.IsGitRefNotFound(cause) {
		return fmt.Errorf("unable to resolve commit for %s@%s: %w", repo, ref, cause)
	}

	data, err := compositeerrors.New(
		&vcserrors.GitRefNotFoundError{Repo: repo, Ref: ref},
		compositeerrors.WithSource("component_builds", buildID),
	)
	if err != nil {
		return fmt.Errorf("unable to build component build composite error: %w", err)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.ComponentBuild{ID: buildID}).
		Select("composite_error", "status", "status_description").
		Updates(app.ComponentBuild{
			CompositeError:    data,
			Status:            app.ComponentBuildStatusError,
			StatusDescription: data.Message,
		}); res.Error != nil {
		return fmt.Errorf("unable to set component build composite error: %w", res.Error)
	}

	return vcserrors.NewGitRefNotFound(repo, ref, cause)
}

func buildRepoRef(connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig) (repo, branch string) {
	if connected != nil {
		return connected.Repo, connected.Branch
	}
	if public != nil {
		return public.Repo, public.Branch
	}
	return "", ""
}

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
