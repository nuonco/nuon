package activities

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FindReusableSandboxBuildInput struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
	RunID       string `json:"run_id" validate:"required"`
}

type FindReusableSandboxBuildOutput struct {
	BuildID string `json:"build_id,omitempty"`
}

func (a *Activities) FindReusableSandboxBuild(ctx context.Context, input *FindReusableSandboxBuildInput) (*FindReusableSandboxBuildOutput, error) {
	out := &FindReusableSandboxBuildOutput{}
	if err := a.v.Struct(input); err != nil {
		return out, nil
	}

	cfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.AppConfigID)
	if err != nil {
		return out, nil
	}

	run, err := a.getAppBranchRunByID(ctx, input.RunID)
	if err != nil {
		return nil, err
	}

	sandboxRepo, sandboxVCSID := sandboxConfigRepo(cfg)
	branchRepo := ""
	if run.AppBranchConfig.ConnectedGithubVCSConfig != nil {
		branchRepo = run.AppBranchConfig.ConnectedGithubVCSConfig.Repo
	} else if run.AppBranchConfig.PublicGitVCSConfig != nil {
		branchRepo = run.AppBranchConfig.PublicGitVCSConfig.Repo
	}
	directory := sandboxConfigDirectory(cfg)

	if repoURLsEqual(sandboxRepo, branchRepo) {
		paths, ok, pathErr := a.comparisonChangedPaths(ctx, input.RunID)
		if pathErr != nil {
			return nil, pathErr
		}
		if !ok || anyPathMatchesDirectory(paths, directory) {
			return out, nil
		}
		build, findErr := a.latestEquivalentSandboxBuild(ctx, input.AppID, cfg)
		if findErr != nil {
			return nil, findErr
		}
		if build != nil {
			out.BuildID = build.ID
		}
		return out, nil
	}

	if sandboxVCSID == "" {
		return nil, fmt.Errorf("sandbox source %s has no vcs config", sandboxRepo)
	}
	latest, err := a.latestVCSCommitSHA(ctx, sandboxVCSID, sandboxConfigBranch(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to resolve sandbox commit: %w", err)
	}
	build, err := a.latestEquivalentSandboxBuild(ctx, input.AppID, cfg)
	if err != nil {
		return nil, err
	}
	if build == nil || build.VCSConnectionCommit == nil || build.VCSConnectionCommit.SHA == "" {
		return out, nil
	}
	previous := build.VCSConnectionCommit.SHA
	if previous != latest {
		paths, pathErr := a.changedPathsBetween(ctx, sandboxVCSID, previous, latest)
		if pathErr != nil {
			return nil, pathErr
		}
		if anyPathMatchesDirectory(paths, directory) {
			return out, nil
		}
	}
	out.BuildID = build.ID
	return out, nil
}

func (a *Activities) latestEquivalentSandboxBuild(ctx context.Context, appID string, cfg *app.AppSandboxConfig) (*app.AppSandboxBuild, error) {
	var builds []app.AppSandboxBuild
	err := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  appID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		Limit(25).
		Find(&builds).Error
	if err != nil {
		return nil, fmt.Errorf("unable to list sandbox builds: %w", err)
	}
	for i := range builds {
		if sandboxConfigsEquivalent(cfg, &builds[i].AppSandboxConfig) {
			build := builds[i]
			return &build, nil
		}
	}
	return nil, nil
}

func sandboxConfigsEquivalent(a, b *app.AppSandboxConfig) bool {
	if a == nil || b == nil {
		return false
	}
	aRepo, _ := sandboxConfigRepo(a)
	bRepo, _ := sandboxConfigRepo(b)
	if !repoURLsEqual(aRepo, bRepo) || sandboxConfigBranch(a) != sandboxConfigBranch(b) {
		return false
	}
	if normalizeRepoPath(sandboxConfigDirectory(a)) != normalizeRepoPath(sandboxConfigDirectory(b)) {
		return false
	}
	if a.TerraformVersion != b.TerraformVersion {
		return false
	}
	return hstoreCanon(a.Variables) == hstoreCanon(b.Variables) && stringListCanon(a.VariablesFiles) == stringListCanon(b.VariablesFiles)
}

func sandboxConfigDirectory(cfg *app.AppSandboxConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.ConnectedGithubVCSConfig != nil {
		return cfg.ConnectedGithubVCSConfig.Directory
	}
	if cfg.PublicGitVCSConfig != nil {
		return cfg.PublicGitVCSConfig.Directory
	}
	return ""
}

func hstoreCanon(h pgtype.Hstore) string {
	if len(h) == 0 {
		return ""
	}
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, key := range keys {
		b.WriteString(key)
		b.WriteByte('=')
		if h[key] != nil {
			b.WriteString(*h[key])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func stringListCanon(values pq.StringArray) string {
	if len(values) == 0 {
		return ""
	}
	copied := append(pq.StringArray(nil), values...)
	sort.Strings(copied)
	return strings.Join(copied, "\n")
}

func AwaitFindReusableSandboxBuild(ctx workflow.Context, input *FindReusableSandboxBuildInput, opts ...*workflow.ActivityOptions) (*FindReusableSandboxBuildOutput, error) {
	var result *FindReusableSandboxBuildOutput
	options := workflow.GetActivityOptions(ctx)
	options.StartToCloseTimeout = time.Minute
	options.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: 3}
	for _, opt := range opts {
		if opt != nil && opt.StartToCloseTimeout != 0 {
			options.StartToCloseTimeout = opt.StartToCloseTimeout
		}
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	err := workflow.ExecuteActivity(ctx, (*Activities).FindReusableSandboxBuild, input).Get(ctx, &result)
	return result, err
}
