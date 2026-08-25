package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	mcpLogPageSize    = 100
	mcpLogMaxPageSize = 200
	mcpLogBodyMaxLen  = 500
)

type mcpLogEntry struct {
	Timestamp   string `json:"timestamp"`
	Severity    string `json:"severity"`
	ServiceName string `json:"service_name,omitempty"`
	Body        string `json:"body"`
}

type mcpLogFilters struct {
	Severity     string
	BodyContains string
	Cursor       string
	Limit        int
}

type mcpLogPage struct {
	LogStreamID     string        `json:"log_stream_id"`
	ReturnedRecords int           `json:"returned_records"`
	HasMore         bool          `json:"has_more"`
	NextCursor      string        `json:"next_cursor,omitempty"`
	Logs            []mcpLogEntry `json:"logs"`
}

func (s *service) resolveOwnerLogStream(ctx context.Context, orgID, ownerType, ownerID string) (string, error) {
	var ls app.LogStream
	err := s.db.WithContext(ctx).
		Where(app.LogStream{
			OrgID:     orgID,
			OwnerType: ownerType,
			OwnerID:   ownerID,
		}).
		First(&ls).Error
	if err != nil {
		return "", fmt.Errorf("no log stream found for %s %s: %w", ownerType, ownerID, err)
	}
	return ls.ID, nil
}

func (s *service) queryMCPLogs(ctx context.Context, orgID, logStreamID string, filters mcpLogFilters) (mcpLogPage, error) {
	limit := mcpLogPageSize
	if filters.Limit > 0 && filters.Limit <= mcpLogMaxPageSize {
		limit = filters.Limit
	}

	cursor, err := parseTailCursor(filters.Cursor)
	if err != nil {
		return mcpLogPage{}, fmt.Errorf("invalid cursor: %w", err)
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

	if filters.Severity != "" {
		q = q.Where("severity_text = ?", filters.Severity)
	}
	if filters.BodyContains != "" {
		q = q.Where("body LIKE ?", "%"+filters.BodyContains+"%")
	}

	var rows []app.OtelLogRecord
	if err := q.Order("timestamp DESC, id DESC").
		Limit(limit + 1).
		Find(&rows).Error; err != nil {
		return mcpLogPage{}, fmt.Errorf("unable to query logs: %w", err)
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

	page := mcpLogPage{
		LogStreamID:     logStreamID,
		ReturnedRecords: len(logs),
		HasMore:         hasMore,
		Logs:            logs,
	}

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		page.NextCursor = encodeTailCursor(tailCursor{
			tsNano: last.Timestamp.UnixNano(),
			id:     last.ID,
		})
	}

	return page, nil
}
