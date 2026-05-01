package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) AccountAuditLogsTable(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")
	page := getPageFromQuery(c)

	// Parse date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day (23:59:59) to include all entries from that day
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	// Parse entity type filters
	var entityTypes []string
	if typeFilter := c.Query("entity_types"); typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				entityTypes = append(entityTypes, trimmed)
			}
		}
	}

	auditLogs, auditLogsTotalPages, err := s.getAuditLogsForAccount(
		ctx, accountID, startDate, endDate, page, entityTypes,
	)
	if err != nil {
		s.l.Error("failed to get audit logs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id":  accountID,
		"audit_logs":  auditLogs,
		"start_date":  startDate,
		"end_date":    endDate,
		"page":        page,
		"total_pages": auditLogsTotalPages,
	})
}
