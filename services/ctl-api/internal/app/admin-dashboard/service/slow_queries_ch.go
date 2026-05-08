package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// QueriesClickHouse returns aggregated query data from the ClickHouse queries table.
func (s *service) QueriesClickHouse(c *gin.Context) {
	search := c.Query("search")
	table := c.Query("table")
	dbType := c.Query("db_type")
	source := c.Query("source")
	sortBy := c.DefaultQuery("sort", "max_ms")
	minDurationMS := c.Query("min_duration_ms")
	timeRange := c.DefaultQuery("time_range", "1h")

	var interval string
	switch timeRange {
	case "5m":
		interval = "5 MINUTE"
	case "15m":
		interval = "15 MINUTE"
	case "30m":
		interval = "30 MINUTE"
	case "1h":
		interval = "1 HOUR"
	case "6h":
		interval = "6 HOUR"
	case "24h":
		interval = "24 HOUR"
	case "7d":
		interval = "7 DAY"
	default:
		interval = "1 HOUR"
	}

	var orderBy string
	switch sortBy {
	case "avg_ms":
		orderBy = "avg_ms DESC"
	case "count":
		orderBy = "count DESC"
	case "total_ms":
		orderBy = "total_ms DESC"
	case "last_seen":
		orderBy = "last_seen_at DESC"
	default:
		orderBy = "max_ms DESC"
	}

	query := s.chDB.WithContext(c).
		Table("queries").
		Select(`
			sql,
			any(` + "`table`" + `) as ` + "`table`" + `,
			any(operation) as operation,
			any(db_type) as db_type,
			any(source) as source,
			any(endpoint) as endpoint,
			count() as count,
			sum(duration_ms) as total_ms,
			avg(duration_ms) as avg_ms,
			min(duration_ms) as min_ms,
			max(duration_ms) as max_ms,
			sum(rows_affected) as total_rows,
			max(rows_affected) as max_rows,
			max(response_size) as max_response_size,
			anyLast(error) as last_error,
			any(caller) as caller,
			any(caller_url) as caller_url,
			formatDateTime(max(timestamp), '%Y-%m-%dT%H:%i:%S.000+00:00') as last_seen_at
		`).
		Where("timestamp >= now() - INTERVAL " + interval).
		Group("sql").
		Order(orderBy).
		Limit(500)

	if table != "" {
		query = query.Where("`table` = ?", table)
	}
	if dbType != "" {
		query = query.Where("db_type = ?", dbType)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("sql ILIKE ? OR `table` ILIKE ? OR caller ILIKE ? OR endpoint ILIKE ?", like, like, like, like)
	}
	if minDurationMS != "" {
		query = query.Having("max_ms >= ?", minDurationMS)
	}

	var results []AggregatedQuery
	if err := query.Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch distinct tables for the filter dropdown.
	var tables []string
	s.chDB.WithContext(c).
		Table("queries").
		Select("DISTINCT `table`").
		Where("timestamp >= now() - INTERVAL "+interval).
		Order("`table`").
		Pluck("`table`", &tables)

	// Total row count in the time range.
	var total int64
	s.chDB.WithContext(c).
		Table("queries").
		Where("timestamp >= now() - INTERVAL " + interval).
		Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"queries": results,
		"tables":  tables,
		"total":   total,
	})
}
