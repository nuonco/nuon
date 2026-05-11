package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// QueriesClickHouse returns aggregated query data from the ClickHouse queries table.
func (s *service) QueriesClickHouse(c *gin.Context) {
	ctx := c.Request.Context()

	search := c.Query("search")
	table := c.Query("table")
	dbType := c.Query("db_type")
	source := c.Query("source")
	hasError := c.Query("has_error")
	sortBy := c.DefaultQuery("sort", "max_ms")
	minDurationMS := c.Query("min_duration_ms")
	timeRange := c.DefaultQuery("time_range", "15m")

	var interval string
	switch timeRange {
	case "15m":
		interval = "15 MINUTE"
	case "1h":
		interval = "1 HOUR"
	case "12h":
		interval = "12 HOUR"
	case "24h":
		interval = "24 HOUR"
	default:
		interval = "15 MINUTE"
	}

	var orderBy string
	switch sortBy {
	case "avg_ms":
		orderBy = "avg_ms DESC"
	case "count":
		orderBy = "cnt DESC"
	case "total_ms":
		orderBy = "total_ms DESC"
	case "last_seen":
		orderBy = "last_seen_at DESC"
	default:
		orderBy = "max_ms DESC"
	}

	// Build query with manual SQL to avoid GORM rewriting ClickHouse-specific syntax.
	var wheres []string
	var args []interface{}

	wheres = append(wheres, "timestamp >= now() - INTERVAL "+interval)

	if table != "" {
		wheres = append(wheres, "`table` = ?")
		args = append(args, table)
	}
	if dbType != "" {
		wheres = append(wheres, "db_type = ?")
		args = append(args, dbType)
	}
	if source != "" {
		wheres = append(wheres, "source = ?")
		args = append(args, source)
	}
	if hasError == "true" {
		wheres = append(wheres, "error != ''")
	}
	if search != "" {
		wheres = append(wheres, "(sql ILIKE ? OR `table` ILIKE ? OR caller ILIKE ? OR endpoint ILIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}

	whereClause := strings.Join(wheres, " AND ")

	havingClause := ""
	if minDurationMS != "" {
		havingClause = "HAVING max_ms >= ?"
		args = append(args, minDurationMS)
	}

	rawSQL := fmt.Sprintf(`SELECT
		sql,
		any(`+"`table`"+`) as agg_table,
		any(operation) as agg_operation,
		any(db_type) as agg_db_type,
		any(source) as agg_source,
		any(endpoint) as agg_endpoint,
		count(*) as cnt,
		sum(duration_ms) as total_ms,
		avg(duration_ms) as avg_ms,
		min(duration_ms) as min_ms,
		max(duration_ms) as max_ms,
		sum(rows_affected) as total_rows,
		max(rows_affected) as max_rows,
		max(response_size) as max_response_size,
		anyLast(error) as agg_last_error,
		any(caller) as agg_caller,
		any(caller_url) as agg_caller_url,
		formatDateTime(max(timestamp), '%%Y-%%m-%%dT%%H:%%i:%%S.000+00:00') as last_seen_at
	FROM queries
	WHERE %s
	GROUP BY sql
	%s
	ORDER BY %s
	LIMIT 500`, whereClause, havingClause, orderBy)

	var results []AggregatedQuery
	if err := s.chDB.WithContext(ctx).Raw(rawSQL, args...).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch distinct tables for the filter dropdown.
	type tblRow struct {
		Tbl string `gorm:"column:tbl"`
	}
	var tableRows []tblRow
	s.chDB.WithContext(ctx).Raw(
		"SELECT DISTINCT `table` as tbl FROM queries WHERE timestamp >= now() - INTERVAL " + interval + " ORDER BY tbl",
	).Scan(&tableRows)
	tables := make([]string, len(tableRows))
	for i, r := range tableRows {
		tables[i] = r.Tbl
	}

	// Total row count in the time range.
	var total int64
	s.chDB.WithContext(ctx).Raw(
		"SELECT count(*) FROM queries WHERE timestamp >= now() - INTERVAL " + interval,
	).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"queries": results,
		"tables":  tables,
		"total":   total,
	})
}
