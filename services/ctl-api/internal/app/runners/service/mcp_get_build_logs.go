package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetBuildLogsInput struct {
	BuildID      string `json:"build_id" jsonschema:"build ID to get logs for"`
	Severity     string `json:"severity,omitempty" jsonschema:"filter by severity (e.g. ERROR, WARN, INFO)"`
	BodyContains string `json:"body_contains,omitempty" jsonschema:"substring filter on log body"`
	Cursor       string `json:"cursor,omitempty" jsonschema:"pagination cursor from a previous response's next_cursor"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max records to return (default 100, max 200)"`
}

type mcpBuildLogResult struct {
	BuildID         string        `json:"build_id"`
	LogStreamID     string        `json:"log_stream_id"`
	ReturnedRecords int           `json:"returned_records"`
	HasMore         bool          `json:"has_more"`
	NextCursor      string        `json:"next_cursor,omitempty"`
	Logs            []mcpLogEntry `json:"logs"`
}

func (s *service) mcpGetBuildLogs(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetBuildLogsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var build app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Where(app.ComponentBuild{OrgID: orgID}).
		Where("id = ?", in.BuildID).
		First(&build).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find build %q: %w", in.BuildID, err)
	}

	logStreamID, err := s.resolveOwnerLogStream(ctx, orgID, "component_builds", build.ID)
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

	return apiPkg.MCPJSONResult(mcpBuildLogResult{
		BuildID:         build.ID,
		LogStreamID:     page.LogStreamID,
		ReturnedRecords: page.ReturnedRecords,
		HasMore:         page.HasMore,
		NextCursor:      page.NextCursor,
		Logs:            page.Logs,
	})
}
