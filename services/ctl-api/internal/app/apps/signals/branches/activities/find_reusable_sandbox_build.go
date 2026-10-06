package activities

import (
	"context"
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

	sha, err := a.sandboxSourceSHA(ctx, input.RunID, cfg)
	if err != nil || sha == "" {
		return out, nil
	}

	var builds []app.AppSandboxBuild
	err = a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  input.AppID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		Limit(25).
		Find(&builds).Error
	if err != nil {
		return out, nil
	}

	for i := range builds {
		build := &builds[i]
		if build.VCSConnectionCommit == nil || build.VCSConnectionCommit.SHA != sha {
			continue
		}
		if sandboxConfigsEquivalent(cfg, &build.AppSandboxConfig) {
			out.BuildID = build.ID
			return out, nil
		}
	}
	return out, nil
}

func (a *Activities) sandboxSourceSHA(ctx context.Context, runID string, cfg *app.AppSandboxConfig) (string, error) {
	run, err := a.getAppBranchRunByID(ctx, runID)
	if err != nil {
		return "", err
	}

	sandboxRepo, sandboxVCSID := sandboxConfigRepo(cfg)
	branchRepo := ""
	if run.AppBranchConfig.ConnectedGithubVCSConfig != nil {
		branchRepo = run.AppBranchConfig.ConnectedGithubVCSConfig.Repo
	} else if run.AppBranchConfig.PublicGitVCSConfig != nil {
		branchRepo = run.AppBranchConfig.PublicGitVCSConfig.Repo
	}
	if repoURLsEqual(sandboxRepo, branchRepo) {
		if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
			return run.VCSConnectionCommit.SHA, nil
		}
		return run.HeadSHA, nil
	}

	gitSource, err := a.GetSandboxBuildGitSource(ctx, GetSandboxBuildGitSourceRequest{
		SandboxConfigID: cfg.ID,
	})
	if err != nil || gitSource == nil || gitSource.Ref == "" || sandboxVCSID == "" {
		return "", err
	}
	commit, err := a.FetchCommitBySHA(ctx, &FetchCommitBySHAInput{
		VcsConfigID: sandboxVCSID,
		SHA:         gitSource.Ref,
	})
	if err != nil || commit == nil {
		return "", err
	}
	return commit.SHA, nil
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
