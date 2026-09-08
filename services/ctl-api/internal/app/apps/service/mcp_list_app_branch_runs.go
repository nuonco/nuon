package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppBranchRunsInput struct {
	App    string `json:"app" jsonschema:"app name or ID"`
	Branch string `json:"branch" jsonschema:"app branch name or ID"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum runs to return (default 20, max 100)"`
}

type mcpAppBranchRunHistoryItem struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	Succeeded        bool   `json:"succeeded"`
	AwaitingApproval bool   `json:"awaiting_approval"`
	RunType          string `json:"run_type"`
	Preview          bool   `json:"preview"`
	PlanOnly         bool   `json:"plan_only"`
	PRNumber         *int   `json:"pr_number,omitempty"`
	HeadSHA          string `json:"head_sha,omitempty"`
	WorkflowID       string `json:"workflow_id,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type mcpListAppBranchRunsResult struct {
	App    mcpAppRef                    `json:"app"`
	Branch mcpAppBranchOverview         `json:"branch"`
	Runs   []mcpAppBranchRunHistoryItem `json:"runs"`
}

func (s *service) mcpListAppBranchRuns(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAppBranchRunsInput) (*mcp.CallToolResult, any, error) {
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

	limit := in.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 100")
	}

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}
	branch, err := s.findAppBranch(ctx, orgID, a.ID, in.Branch)
	if err != nil {
		return nil, nil, err
	}

	var runs []app.AppBranchRun
	res := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("Preview").
		Where(app.AppBranchRun{AppBranchID: branch.ID}).
		Order("created_at DESC").
		Limit(limit).
		Find(&runs)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list app branch runs: %w", res.Error)
	}

	result := mcpListAppBranchRunsResult{
		App:    mcpAppRef{ID: a.ID, Name: a.Name},
		Branch: mcpAppBranchOverview{ID: branch.ID, Name: branch.Name, ManagedBy: string(branch.ManagedBy)},
		Runs:   make([]mcpAppBranchRunHistoryItem, 0, len(runs)),
	}
	for i := range runs {
		run := &runs[i]
		if err := s.markRunAwaitingApproval(ctx, run); err != nil {
			return nil, nil, err
		}
		headSHA := run.HeadSHA
		if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
			headSHA = run.VCSConnectionCommit.SHA
		}
		workflowID := ""
		if run.WorkflowID != nil {
			workflowID = *run.WorkflowID
		}
		result.Runs = append(result.Runs, mcpAppBranchRunHistoryItem{
			ID:               run.ID,
			Status:           run.Status,
			Succeeded:        run.Status == "success",
			AwaitingApproval: run.AwaitingApproval,
			RunType:          string(run.RunType),
			Preview:          run.IsPreview(),
			PlanOnly:         run.PlanOnly,
			PRNumber:         run.PRNumber,
			HeadSHA:          headSHA,
			WorkflowID:       workflowID,
			CreatedAt:        apiPkg.MCPTime(run.CreatedAt),
		})
	}

	return apiPkg.MCPJSONResult(result)
}
