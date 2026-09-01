package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppBranchPreviewSourcesInput struct {
	App    string `json:"app" jsonschema:"app name or ID"`
	Branch string `json:"branch" jsonschema:"app branch name or ID"`
}

func (s *service) mcpListAppBranchPreviewSources(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAppBranchPreviewSourcesInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := s.requireAppBranches(ctx); err != nil {
		return nil, nil, err
	}
	if in.App == "" {
		return nil, nil, fmt.Errorf("app is required")
	}
	if in.Branch == "" {
		return nil, nil, fmt.Errorf("branch is required")
	}

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}
	branch, err := s.findAppBranch(ctx, orgID, a.ID, in.Branch)
	if err != nil {
		return nil, nil, err
	}

	var config app.AppBranchConfig
	res := s.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig.VCSConnection").
		Preload("PublicGitVCSConfig").
		Where("app_branch_id = ?", branch.ID).
		Order("config_number DESC").
		First(&config)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to find branch config: %w", res.Error)
	}

	sources, err := s.helpers.ListPreviewSources(ctx, branch, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list preview sources: %w", err)
	}

	return apiPkg.MCPJSONResult(sources)
}
