package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const logStreamLogsPerPage = 100

// LogStreamViewer renders the log stream search page
func (s *service) LogStreamViewer(c *gin.Context) {
	component := views.LogStreamSearch()
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// LogStreamDetail renders the log stream detail page with logs
func (s *service) LogStreamDetail(c *gin.Context) {
	ctx := c.Request.Context()
	logStreamID := c.Param("log_stream_id")
	page := getPageFromQuery(c)

	// Find log stream by ID or owner ID
	ls, err := s.getLogStream(ctx, logStreamID)
	if err != nil {
		s.l.Error("failed to fetch log stream", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Log stream not found"})
		return
	}

	// Fetch logs from ClickHouse
	logs, totalPages, err := s.getLogStreamLogs(ctx, ls.ID, ls.OrgID, page)
	if err != nil {
		s.l.Warn("failed to fetch log stream logs", zap.Error(err))
		logs = []app.OtelLogRecord{}
		totalPages = 1
	}

	component := views.LogStreamDetail(ls, logs, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// LogStreamLogsTable handles the HTMX endpoint for log pagination
func (s *service) LogStreamLogsTable(c *gin.Context) {
	ctx := c.Request.Context()
	logStreamID := c.Param("log_stream_id")
	page := getPageFromQuery(c)

	ls, err := s.getLogStream(ctx, logStreamID)
	if err != nil {
		s.l.Error("failed to fetch log stream", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Log stream not found"})
		return
	}

	logs, totalPages, err := s.getLogStreamLogs(ctx, ls.ID, ls.OrgID, page)
	if err != nil {
		s.l.Error("failed to fetch logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	component := views.LogStreamLogsTable(ls.ID, logs, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getLogStream(ctx context.Context, logStreamID string) (*app.LogStream, error) {
	var ls app.LogStream
	res := s.db.WithContext(ctx).
		Where("id = ?", logStreamID).
		Or("owner_id = ?", logStreamID).
		First(&ls)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get log stream: %w", res.Error)
	}
	return &ls, nil
}

func (s *service) getLogStreamLogs(ctx context.Context, logStreamID, orgID string, page int) ([]app.OtelLogRecord, int, error) {
	ctx, cancelFn := context.WithTimeout(ctx, time.Second*10)
	defer cancelFn()

	// Count total logs
	var totalCount int64
	countRes := s.chDB.WithContext(ctx).
		Model(&app.OtelLogRecord{}).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID).
		Count(&totalCount)
	if countRes.Error != nil {
		return nil, 0, fmt.Errorf("unable to count logs: %w", countRes.Error)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(logStreamLogsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	if totalCount == 0 {
		return []app.OtelLogRecord{}, totalPages, nil
	}

	offset := (page - 1) * logStreamLogsPerPage
	var logs []app.OtelLogRecord
	res := s.chDB.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("log_stream_id = ?", logStreamID).
		Order("timestamp ASC").
		Offset(offset).
		Limit(logStreamLogsPerPage).
		Find(&logs)
	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to retrieve logs: %w", res.Error)
	}

	return logs, totalPages, nil
}

// helper to format log count
func formatLogCount(count int) string {
	if count == 0 {
		return "0"
	}
	return strconv.Itoa(count)
}
