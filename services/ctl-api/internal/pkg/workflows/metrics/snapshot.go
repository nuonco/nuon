package workflowmetrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

var workflowTypes = []string{"provision", "manual_deploy", "deploy_components"}
var workflowStates = []string{"queued", "executing", "awaiting_approval", "awaiting_retry"}

type bucket struct {
	workflowType string
	state        string
}

type value struct {
	count  int64
	oldest time.Time
}

// FinishedAt can be set before a retryable park. Retry states instead require
// a nonterminal execute signal; workflow metadata can outlive that signal.
// Materialize the candidate states to avoid repeating signal lookups for grouping.
const snapshotQuery = `
WITH current_workflows AS MATERIALIZED (
  SELECT w.type, w.created_at,
    CASE
      WHEN w.status->>'status' IN ('pending', 'queued') THEN 'queued'
      WHEN w.status->>'status' IN ('in-progress', 'retrying') THEN 'executing'
      WHEN w.status->>'status' = 'approval-awaiting' THEN 'awaiting_approval'
      WHEN (w.status->>'status' = 'failed-pending-retry'
        OR (w.status->>'status' = 'error'
          AND w.status->'metadata'->>'awaiting_retry' = 'true'))
        AND COALESCE(w.status->'metadata'->>'stopped', 'false') <> 'true'
        AND COALESCE(w.status->'metadata'->>'retries_exhausted', 'false') <> 'true'
        AND EXISTS (
          SELECT 1 FROM queue_signals qs
          WHERE qs.owner_id = w.id AND qs.owner_type = 'install_workflows'
            AND qs.type = 'execute-workflow' AND qs.deleted_at = 0
            AND qs.status->>'status' NOT IN ('success', 'error', 'cancelled')
        )
        THEN 'awaiting_retry'
    END AS state
  FROM install_workflows w
  JOIN installs i ON i.id = w.owner_id AND i.deleted_at = 0
  WHERE w.owner_type = 'installs' AND w.deleted_at = 0
    AND w.type = ANY($1) AND w.plan_only IS NOT TRUE
    AND (w.status->>'status' IN ('pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'failed-pending-retry')
      OR (w.status->>'status' = 'error' AND w.status->'metadata'->>'awaiting_retry' = 'true'))
)
SELECT type, state, count(*), min(created_at)
FROM current_workflows WHERE state IS NOT NULL
GROUP BY type, state`

func readSnapshot(ctx context.Context, conn *pgx.Conn) (map[bucket]value, error) {
	rows, err := conn.Query(ctx, snapshotQuery, workflowTypes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[bucket]value)
	for rows.Next() {
		var key bucket
		var v value
		if err := rows.Scan(&key.workflowType, &key.state, &v.count, &v.oldest); err != nil {
			return nil, err
		}
		result[key] = v
	}
	return result, rows.Err()
}
