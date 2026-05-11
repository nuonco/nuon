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

// queryCatalog is grouped by purpose:
//   - Storage:       table-sizes, bloat-dead-tuples
//   - Indexes:       index-usage, unused-indexes, fk-missing-indexes, seq-scans
//   - Cache:         cache-hit-ratio
//   - Live sessions: connection-counts, long-running-queries, idle-in-transaction,
//     active-locks, blocked-queries
//   - Queue signals: unenqueued-signals, unenqueued-signals-count
var queryCatalog = []CatalogQuery{
	// --- Storage ---
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
		ID:          "bloat-dead-tuples",
		Name:        "Bloat / dead tuples per table",
		Description: "Top 20 tables by dead-tuple count. High dead_pct with stale last_autovacuum points to autovacuum lagging or being blocked.",
		DBType:      "psql",
		SQL: `SELECT
  schemaname,
  relname,
  n_live_tup,
  n_dead_tup,
  round(n_dead_tup::numeric / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_pct,
  last_vacuum,
  last_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 0
ORDER BY n_dead_tup DESC
LIMIT 20`,
	},

	// --- Indexes ---
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
		ID:          "fk-missing-indexes",
		Name:        "FK columns missing indexes",
		Description: "Foreign keys without a supporting index. Postgres does not auto-create these; missing ones cause slow DELETE/UPDATE on parent tables.",
		DBType:      "psql",
		SQL: `SELECT
  c.conrelid::regclass AS table_name,
  c.conname AS constraint_name,
  pg_get_constraintdef(c.oid) AS constraint_def,
  pg_class.reltuples::bigint AS approx_rows
FROM pg_constraint c
JOIN pg_class ON pg_class.oid = c.conrelid
WHERE c.contype = 'f'
  AND NOT EXISTS (
    SELECT 1 FROM pg_index i
    WHERE i.indrelid = c.conrelid
      AND c.conkey = (i.indkey::smallint[])[1:array_length(c.conkey, 1)]
  )
ORDER BY pg_class.reltuples DESC`,
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

	// --- Cache ---
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

	// --- Live sessions ---
	{
		ID:          "connection-counts",
		Name:        "Connection counts by state",
		Description: "Aggregate connection counts grouped by state. Quick way to spot connection storms or a pile of idle-in-transaction sessions.",
		DBType:      "psql",
		SQL: `SELECT
  state,
  count(*) AS count
FROM pg_stat_activity
GROUP BY state
ORDER BY count DESC`,
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
		ID:          "idle-in-transaction",
		Name:        "Idle-in-transaction sessions",
		Description: "Sessions stuck in 'idle in transaction'. These hold locks and block VACUUM. Longest-idle first.",
		DBType:      "psql",
		SQL: `SELECT
  pid,
  usename,
  state,
  age(now(), state_change) AS idle_for,
  query
FROM pg_stat_activity
WHERE state = 'idle in transaction'
ORDER BY age(now(), state_change) DESC`,
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
		ID:          "blocked-queries",
		Name:        "Blocked queries and blockers",
		Description: "Snapshot of currently blocked queries and the PIDs blocking them. Run during an incident — blocking_pid is the PID to investigate or terminate.",
		DBType:      "psql",
		SQL: `SELECT
  activity.pid,
  activity.usename,
  activity.query,
  activity.wait_event_type,
  activity.wait_event,
  blocking.pid AS blocking_pid,
  blocking.query AS blocking_query
FROM pg_stat_activity AS activity
JOIN pg_stat_activity AS blocking
  ON blocking.pid = ANY(pg_blocking_pids(activity.pid))`,
	},

	// --- Queue signals ---
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
