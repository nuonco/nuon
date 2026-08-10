package service

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

const (
	mcpLogPageSize    = 100
	mcpLogMaxPageSize = 200
	mcpLogBodyMaxLen  = 500
)

type mcpGetWorkflowStepLogsInput struct {
	StepID       string `json:"step_id" jsonschema:"workflow step ID to get logs for"`
	Severity     string `json:"severity,omitempty" jsonschema:"filter by severity (e.g. ERROR, WARN, INFO)"`
	BodyContains string `json:"body_contains,omitempty" jsonschema:"substring filter on log body"`
	Cursor       string `json:"cursor,omitempty" jsonschema:"pagination cursor from a previous response's next_cursor"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max records to return (default 100, max 200)"`
}

type mcpLogEntry struct {
	Timestamp   string `json:"timestamp"`
	Severity    string `json:"severity"`
	ServiceName string `json:"service_name,omitempty"`
	Body        string `json:"body"`
}

type mcpLogResult struct {
	StepName        string        `json:"step_name"`
	LogStreamID     string        `json:"log_stream_id"`
	ReturnedRecords int           `json:"returned_records"`
	HasMore         bool          `json:"has_more"`
	NextCursor      string        `json:"next_cursor,omitempty"`
	Logs            []mcpLogEntry `json:"logs"`
}

func (s *service) mcpGetWorkflowStepLogs(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetWorkflowStepLogsInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where(app.WorkflowStep{OrgID: orgID}).
		Where("id = ?", in.StepID).
		First(&step).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find step %q: %w", in.StepID, err)
	}

	if step.StepTargetID == "" {
		return nil, nil, fmt.Errorf("step %q does not have a target", in.StepID)
	}

	logStreamID, err := s.resolveStepLogStream(ctx, orgID, step.StepTargetType, step.StepTargetID)
	if err != nil {
		return nil, nil, err
	}

	limit := mcpLogPageSize
	if in.Limit > 0 && in.Limit <= mcpLogMaxPageSize {
		limit = in.Limit
	}

	cursor, err := parseTailCursor(in.Cursor)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid cursor: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := s.chDB.WithContext(queryCtx).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID)

	if cursor.tsNano > 0 {
		if cursor.id != "" {
			q = q.Where(
				"(timestamp < fromUnixTimestamp64Nano(?)) OR (timestamp = fromUnixTimestamp64Nano(?) AND id < ?)",
				cursor.tsNano, cursor.tsNano, cursor.id,
			)
		} else {
			q = q.Where("timestamp < fromUnixTimestamp64Nano(?)", cursor.tsNano)
		}
	}

	if in.Severity != "" {
		q = q.Where("severity_text = ?", in.Severity)
	}
	if in.BodyContains != "" {
		q = q.Where("body LIKE ?", "%"+in.BodyContains+"%")
	}

	var rows []app.OtelLogRecord
	if err := q.Order("timestamp DESC, id DESC").
		Limit(limit + 1).
		Find(&rows).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to query logs: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	logs := make([]mcpLogEntry, 0, len(rows))
	for _, r := range rows {
		body := r.Body
		if len(body) > mcpLogBodyMaxLen {
			body = body[:mcpLogBodyMaxLen] + "..."
		}
		logs = append(logs, mcpLogEntry{
			Timestamp:   r.Timestamp.Format(time.RFC3339Nano),
			Severity:    r.SeverityText,
			ServiceName: r.ServiceName,
			Body:        body,
		})
	}

	result := mcpLogResult{
		StepName:        step.Name,
		LogStreamID:     logStreamID,
		ReturnedRecords: len(logs),
		HasMore:         hasMore,
		Logs:            logs,
	}

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		result.NextCursor = encodeTailCursor(tailCursor{
			tsNano: last.Timestamp.UnixNano(),
			id:     last.ID,
		})
	}

	return apiPkg.MCPJSONResult(result)
}

func (s *service) resolveStepLogStream(ctx context.Context, orgID, targetType, targetID string) (string, error) {
	var ownerType string
	switch targetType {
	case string(app.WorkflowStepTargetTypeInstallDeploys):
		ownerType = "install_deploys"
	case string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns):
		ownerType = "install_action_workflow_runs"
	case string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		ownerType = "install_sandbox_runs"
	default:
		return "", fmt.Errorf("step target type %q does not support logs", targetType)
	}

	var ls app.LogStream
	err := s.db.WithContext(ctx).
		Where(app.LogStream{
			OrgID:     orgID,
			OwnerType: ownerType,
			OwnerID:   targetID,
		}).
		First(&ls).Error
	if err != nil {
		return "", fmt.Errorf("no log stream found for %s %s: %w", targetType, targetID, err)
	}
	return ls.ID, nil
}
