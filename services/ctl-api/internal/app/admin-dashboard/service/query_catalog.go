package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CatalogQuery is a pre-defined SQL query that can be run from the admin dashboard.
type CatalogQuery struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
	DBType      string `json:"db_type"` // "psql" or "ch"
}

var queryCatalog = []CatalogQuery{
	{
		ID:          "index-usage",
		Name:        "Index usage (non-unique)",
		Description: "All non-unique, non-primary indexes with scan counts and sizes. Lowest-usage first, biggest within each scan count first.",
		DBType:      "psql",
		SQL: `SELECT
  s.schemaname,
  s.relname AS table_name,
  s.indexrelname AS index_name,
  s.idx_scan,
  pg_size_pretty(pg_relation_size(s.indexrelid)) AS index_size,
  pg_relation_size(s.indexrelid) AS index_size_bytes
FROM pg_stat_user_indexes s
JOIN pg_index i ON s.indexrelid = i.indexrelid
WHERE NOT i.indisunique
  AND NOT i.indisprimary
ORDER BY s.idx_scan ASC, pg_relation_size(s.indexrelid) DESC`,
	},
	{
		ID:          "table-sizes",
		Name:        "Table sizes",
		Description: "Shows all tables ordered by total size (table + indexes).",
		DBType:      "psql",
		SQL: `SELECT
  schemaname,
  relname AS table_name,
  pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
  pg_size_pretty(pg_relation_size(relid)) AS table_size,
  pg_size_pretty(pg_indexes_size(relid)) AS index_size,
  n_live_tup AS row_estimate
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC`,
	},
	{
		ID:          "unused-indexes",
		Name:        "Unused indexes",
		Description: "Non-unique indexes with zero scans. Candidates for removal.",
		DBType:      "psql",
		SQL: `SELECT
  s.schemaname,
  s.relname AS table_name,
  s.indexrelname AS index_name,
  s.idx_scan,
  pg_size_pretty(pg_relation_size(s.indexrelid)) AS index_size
FROM pg_stat_user_indexes s
JOIN pg_index i ON s.indexrelid = i.indexrelid
WHERE NOT i.indisunique
  AND NOT i.indisprimary
  AND s.idx_scan = 0
ORDER BY pg_relation_size(s.indexrelid) DESC`,
	},
	{
		ID:          "active-locks",
		Name:        "Active locks",
		Description: "Shows currently held locks with the associated query.",
		DBType:      "psql",
		SQL: `SELECT
  l.locktype,
  l.relation::regclass AS table_name,
  l.mode,
  l.granted,
  a.usename,
  a.query,
  a.state,
  a.query_start,
  age(now(), a.query_start) AS duration
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE a.query NOT ILIKE '%pg_locks%'
ORDER BY a.query_start`,
	},
	{
		ID:          "long-running-queries",
		Name:        "Long-running queries",
		Description: "Active queries running longer than 5 seconds.",
		DBType:      "psql",
		SQL: `SELECT
  pid,
  usename,
  state,
  query,
  query_start,
  age(now(), query_start) AS duration
FROM pg_stat_activity
WHERE state = 'active'
  AND query_start < now() - interval '5 seconds'
ORDER BY query_start`,
	},
	{
		ID:          "seq-scans",
		Name:        "Sequential scans on large tables",
		Description: "Tables doing seq scans on big data — candidates for missing indexes.",
		DBType:      "psql",
		SQL: `SELECT
  relname,
  seq_scan,
  seq_tup_read,
  idx_scan,
  n_live_tup
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_tup_read DESC
LIMIT 20`,
	},
	{
		ID:          "unenqueued-signals",
		Name:        "Unenqueued queue signals",
		Description: "Queue signals that have not been enqueued yet. Newest first.",
		DBType:      "psql",
		SQL: `SELECT
  id,
  org_id,
  queue_id,
  owner_id,
  owner_type,
  type,
  status,
  enqueued,
  execution_count,
  created_at
FROM queue_signals
WHERE enqueued = false
  AND deleted_at = 0
ORDER BY created_at DESC
LIMIT 200`,
	},
	{
		ID:          "unenqueued-signals-count",
		Name:        "Unenqueued queue signals (count)",
		Description: "Count of un-enqueued queue signals grouped by owner_type and type.",
		DBType:      "psql",
		SQL: `SELECT
  owner_type,
  type,
  COUNT(*) AS count
FROM queue_signals
WHERE enqueued = false
  AND deleted_at = 0
GROUP BY owner_type, type
ORDER BY count DESC`,
	},
	{
		ID:          "cache-hit-ratio",
		Name:        "Cache hit ratio",
		Description: "Buffer cache hit ratio per table. Want >99% for hot tables.",
		DBType:      "psql",
		SQL: `SELECT relname,
       heap_blks_read,
       heap_blks_hit,
       round(heap_blks_hit::numeric / nullif(heap_blks_hit + heap_blks_read, 0), 4) AS hit_ratio
FROM pg_statio_user_tables
ORDER BY heap_blks_read DESC`,
	},
}

func (s *service) QueryCatalogList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"queries": queryCatalog,
	})
}

func (s *service) QueryCatalogRun(c *gin.Context) {
	id := c.Param("query_id")

	var found *CatalogQuery
	for i := range queryCatalog {
		if queryCatalog[i].ID == id {
			found = &queryCatalog[i]
			break
		}
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "query not found"})
		return
	}

	target := c.DefaultQuery("target", "replica")
	db := s.psqlForTarget(target)
	if found.DBType == "ch" {
		db = s.chDB
	}

	var results []map[string]any
	if err := db.WithContext(c.Request.Context()).Raw(found.SQL).Scan(&results).Error; err != nil {
		s.l.Error("catalog query failed", zap.String("query_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   found,
		"target":  target,
		"results": results,
		"count":   len(results),
	})
}

func (s *service) QueryCollectorToggle(c *gin.Context) {
	if s.queryCollector == nil {
		c.JSON(http.StatusOK, gin.H{"enabled": false, "message": "collector not initialized (requires debug_enable_query_collector=true at startup)"})
		return
	}

	action := c.Query("action")
	switch action {
	case "clear":
		s.queryCollector.Clear()
		c.JSON(http.StatusOK, gin.H{"enabled": true, "cleared": true})
	default:
		c.JSON(http.StatusOK, gin.H{"enabled": true, "total": s.queryCollector.Total()})
	}
}
