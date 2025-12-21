package activities

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/workspace"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
// @by-field vcsConfigID
func (a *Activities) fetchIntermediateConfig(ctx context.Context, vcsConfigID string, commitSHA string) (*config.AppConfig, error) {
	gitSource, err := a.resolveGitSource(ctx, vcsConfigID, commitSHA)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve git source: %w", err)
	}

	ws, err := workspace.New(a.v,
		workspace.WithGitSource(gitSource),
		workspace.WithID("app-branch-"+vcsConfigID),
		workspace.WithLogger(zap.L()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}
	defer ws.Cleanup(ctx)

	if err := ws.Init(ctx); err != nil {
		return nil, fmt.Errorf("unable to init workspace: %w", err)
	}

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       ws.SourceDir(),
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse config from repo: %w", err)
	}

	return cfg, nil
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
