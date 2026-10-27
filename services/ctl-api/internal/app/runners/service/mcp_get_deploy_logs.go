package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetDeployLogsInput struct {
	DeployID     string `json:"deploy_id" jsonschema:"deploy ID to get logs for"`
	Severity     string `json:"severity,omitempty" jsonschema:"filter by severity (e.g. ERROR, WARN, INFO)"`
	BodyContains string `json:"body_contains,omitempty" jsonschema:"substring filter on log body"`
	Cursor       string `json:"cursor,omitempty" jsonschema:"pagination cursor from a previous response's next_cursor"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max records to return (default 100, max 200)"`
}

type mcpDeployLogResult struct {
	DeployID        string        `json:"deploy_id"`
	LogStreamID     string        `json:"log_stream_id"`
	ReturnedRecords int           `json:"returned_records"`
	HasMore         bool          `json:"has_more"`
	NextCursor      string        `json:"next_cursor,omitempty"`
	Logs            []mcpLogEntry `json:"logs"`
}

func (s *service) mcpGetDeployLogs(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetDeployLogsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var deploy app.InstallDeploy
	if err := s.db.WithContext(ctx).
		Where(app.InstallDeploy{OrgID: orgID}).
		Where("id = ?", in.DeployID).
		First(&deploy).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find deploy %q: %w", in.DeployID, err)
	}

	logStreamID, err := s.resolveOwnerLogStream(ctx, orgID, "install_deploys", deploy.ID)
	if err != nil {
		return nil, nil, err
	}

	page, err := s.queryMCPLogs(ctx, orgID, logStreamID, mcpLogFilters{
		Severity:     in.Severity,
		BodyContains: in.BodyContains,
		Cursor:       in.Cursor,
		Limit:        in.Limit,
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpDeployLogResult{
		DeployID:        deploy.ID,
		LogStreamID:     page.LogStreamID,
		ReturnedRecords: page.ReturnedRecords,
		HasMore:         page.HasMore,
		NextCursor:      page.NextCursor,
		Logs:            page.Logs,
	})
}
