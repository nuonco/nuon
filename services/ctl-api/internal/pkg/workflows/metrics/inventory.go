package workflowmetrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const inventoryRefreshInterval = 5 * time.Minute
const inventorySnapshotTTL = 15 * time.Minute

var inventoryStates = map[string][]string{
	"org":     {"active", "error", "provisioning", "deleting", "deprovisioning", "deprovisioned", "unknown"},
	"app":     {"active", "error", "provisioning", "deprovisioning", "delete_queued", "unknown"},
	"install": {"provisioning", "provisioned", "deprovisioning", "deprovisioned", "reprovisioning", "unknown"},
	"queue":   {"awaiting_dispatch", "queued", "executing"},
}

type inventoryBucket struct{ kind, state string }

// Materialize live parents once; current queue candidates use the existing
// partial index rather than scanning terminal signal history.
const inventoryQuery = `
WITH live_orgs AS MATERIALIZED (
  SELECT id, COALESCE(NULLIF(status_v2->>'status', ''), status) AS state
  FROM orgs WHERE deleted_at = 0
), live_apps AS MATERIALIZED (
  SELECT a.id, a.org_id, COALESCE(NULLIF(a.status_v2->>'status', ''), a.status) AS state
  FROM apps a JOIN live_orgs o ON o.id = a.org_id WHERE a.deleted_at = 0
), entries AS (
  SELECT 'org' AS kind, state, NULL::timestamptz AS created_at FROM live_orgs
  UNION ALL
  SELECT 'app', state, NULL::timestamptz FROM live_apps
  UNION ALL
  SELECT 'install', i.lifecycle_phase->>'phase', NULL::timestamptz
  FROM installs i JOIN live_apps a ON a.id = i.app_id AND a.org_id = i.org_id
  WHERE i.deleted_at = 0
  UNION ALL
  SELECT 'queue', CASE WHEN qs.status->>'status' = 'in-progress' THEN 'executing'
    WHEN qs.enqueued THEN 'queued' ELSE 'awaiting_dispatch' END, qs.created_at
  FROM queue_signals qs JOIN queues q ON q.id = qs.queue_id AND q.deleted_at = 0
  JOIN live_orgs o ON o.id = qs.org_id
  WHERE qs.deleted_at = 0 AND qs.status->>'status' IN ('queued', 'in-progress')
), classified AS (
  SELECT kind, CASE
    WHEN kind = 'org' AND state IN ('active','error','provisioning','deleting','deprovisioning','deprovisioned') THEN state
    WHEN kind = 'app' AND state IN ('active','error','provisioning','deprovisioning','delete_queued') THEN state
    WHEN kind = 'install' AND state IN ('provisioning','provisioned','deprovisioning','deprovisioned','reprovisioning') THEN state
    WHEN kind = 'queue' THEN state
    ELSE 'unknown' END AS state, created_at
  FROM entries
)
SELECT kind, state, count(*), min(created_at) FROM classified GROUP BY kind, state`

func readInventory(ctx context.Context, conn *pgx.Conn) (map[inventoryBucket]value, error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout='5s'; SET LOCAL lock_timeout='100ms';
SET LOCAL work_mem='4MB'; SET LOCAL max_parallel_workers_per_gather=0`); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, inventoryQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[inventoryBucket]value)
	for rows.Next() {
		var key inventoryBucket
		var v value
		var oldest *time.Time
		if err := rows.Scan(&key.kind, &key.state, &v.count, &oldest); err != nil {
			return nil, err
		}
		if oldest != nil {
			v.oldest = *oldest
		}
		result[key] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}

func (r *reporter) refreshInventoryIfDue(ctx context.Context, conn *pgx.Conn) error {
	if time.Since(r.inventoryAttempt) < inventoryRefreshInterval {
		return nil
	}
	r.inventoryAttempt = time.Now()
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	values, err := readInventory(ctx, conn)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inventory, r.inventoryCollected, r.inventoryActive = values, time.Now(), true
	return nil
}

func (r *reporter) observeInventory(o metric.Observer, current metric.Int64ObservableGauge, queued metric.Int64ObservableGauge, oldest, collected metric.Float64ObservableGauge) {
	if r.inventoryCollected.IsZero() {
		return
	}
	o.ObserveFloat64(collected, float64(r.inventoryCollected.UnixNano())/1e9)
	if !r.inventoryActive || time.Since(r.inventoryCollected) > inventorySnapshotTTL {
		return
	}
	for kind, states := range inventoryStates {
		for _, state := range states {
			v := r.inventory[inventoryBucket{kind, state}]
			if kind == "queue" {
				opts := metric.WithAttributes(attribute.String("queue.state", state))
				o.ObserveInt64(queued, v.count, opts)
				if v.count > 0 {
					o.ObserveFloat64(oldest, float64(v.oldest.UnixNano())/1e9, opts)
				}
			} else {
				o.ObserveInt64(current, v.count, metric.WithAttributes(attribute.String("resource.kind", kind), attribute.String("resource.state", state)))
			}
		}
	}
}
