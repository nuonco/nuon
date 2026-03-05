package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/workspace"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CloneRepoResult contains the result of a repo clone operation.
type CloneRepoResult struct {
	WorkspaceID string `json:"workspace_id"`
	SourceDir   string `json:"source_dir"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) cloneRepo(ctx context.Context, runID string, vcsConfigID string, commitSHA string) (*CloneRepoResult, error) {
	workspaceID := "app-branch-run-" + runID

	gitSource, err := a.resolveGitSource(ctx, vcsConfigID, commitSHA)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve git source: %w", err)
	}

	ws, err := workspace.New(a.v,
		workspace.WithGitSource(gitSource),
		workspace.WithID(workspaceID),
		workspace.WithCleanup(true),
		workspace.WithLogger(zap.L()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	if err := ws.Init(ctx); err != nil {
		return nil, fmt.Errorf("unable to init workspace: %w", err)
	}

	return &CloneRepoResult{
		WorkspaceID: workspaceID,
		SourceDir:   ws.SourceDir(),
	}, nil
}

// resolveGitSource looks up the VCS config by ID and constructs a workspace.GitSource.
// It tries ConnectedGithubVCSConfig first (private repos), then PublicGitVCSConfig (public repos).
func (a *Activities) resolveGitSource(ctx context.Context, vcsConfigID string, commitSHA string) (*workspace.GitSource, error) {
	vcsHelpers := a.helpers.VCSHelpers()

	// Try ConnectedGithubVCSConfig first
	var connectedCfg app.ConnectedGithubVCSConfig
	res := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", vcsConfigID)
	if res.Error == nil {
		src, err := vcsHelpers.GetGitSourceAtCommit(ctx, &connectedCfg, commitSHA)
		if err != nil {
			return nil, fmt.Errorf("unable to get git source for connected repo: %w", err)
		}
		return workspace.GitSourceFromPlanTypes(src), nil
	}

	// Try PublicGitVCSConfig
	var publicCfg app.PublicGitVCSConfig
	res = a.db.WithContext(ctx).First(&publicCfg, "id = ?", vcsConfigID)
	if res.Error == nil {
		src, err := vcsHelpers.GetPublicGitSourceAtCommit(&publicCfg, commitSHA)
		if err != nil {
			return nil, fmt.Errorf("unable to get git source for public repo: %w", err)
		}
		return workspace.GitSourceFromPlanTypes(src), nil
	}

	return nil, fmt.Errorf("VCS config not found: %s", vcsConfigID)
}
