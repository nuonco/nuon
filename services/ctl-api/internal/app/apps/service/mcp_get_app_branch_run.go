package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetAppBranchRunInput struct {
	App      string `json:"app" jsonschema:"app name or ID"`
	Branch   string `json:"branch" jsonschema:"app branch name or ID"`
	RunID    string `json:"run_id,omitempty" jsonschema:"specific app branch run ID; mutually exclusive with pr_number"`
	PRNumber int    `json:"pr_number,omitempty" jsonschema:"newest run for this pull request; mutually exclusive with run_id"`
}

type mcpGetAppBranchRunResult struct {
	App    mcpAppRef               `json:"app"`
	Branch mcpAppBranchOverview    `json:"branch"`
	Run    mcpAppBranchRunOverview `json:"run"`
}

func (s *service) mcpGetAppBranchRun(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppBranchRunInput) (*mcp.CallToolResult, any, error) {
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
	if in.RunID != "" && in.PRNumber > 0 {
		return nil, nil, fmt.Errorf("specify only one of run_id or pr_number")
	}
	if in.PRNumber < 0 {
		return nil, nil, fmt.Errorf("pr_number must be positive")
	}

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}
	branch, err := s.findAppBranch(ctx, orgID, a.ID, in.Branch)
	if err != nil {
		return nil, nil, err
	}

	query := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("Preview").
		Preload("Comparison").
		Preload("Comparison.BaseRun").
		Preload("Comparison.BaseRun.VCSConnectionCommit").
		Where(app.AppBranchRun{AppBranchID: branch.ID})

	switch {
	case in.RunID != "":
		query = query.Where(app.AppBranchRun{ID: in.RunID})
	case in.PRNumber > 0:
		query = query.Where(app.AppBranchRun{PRNumber: &in.PRNumber}).Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var run app.AppBranchRun
	if err := query.First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			switch {
			case in.RunID != "":
				return nil, nil, fmt.Errorf("app branch run %q not found", in.RunID)
			case in.PRNumber > 0:
				return nil, nil, fmt.Errorf("no app branch run found for pull request %d", in.PRNumber)
			default:
				return nil, nil, fmt.Errorf("app branch %q has no runs", branch.Name)
			}
		}
		return nil, nil, fmt.Errorf("unable to get app branch run: %w", err)
	}

	overview, err := s.appBranchRunOverview(ctx, &run)
	if err != nil {
		return nil, nil, err
	}
	result := mcpGetAppBranchRunResult{
		App:    mcpAppRef{ID: a.ID, Name: a.Name},
		Branch: mcpAppBranchOverview{ID: branch.ID, Name: branch.Name, ManagedBy: string(branch.ManagedBy)},
		Run:    *overview,
	}
	return apiPkg.MCPJSONResult(result)
}
