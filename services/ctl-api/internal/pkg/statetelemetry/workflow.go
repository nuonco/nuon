package statetelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var workflowStates = []string{"pending", "queued", "planning", "in-progress", "retrying", "approval-awaiting", "failed-pending-retry", "error", "unset", "other"}

const currentWorkflowsSQL = `
, workflow_records AS (
    SELECT i.org_id, i.app_id, i.id AS install_id, w.type, w.plan_only, w.created_at, w.finished_at,
        COALESCE(w.status->>'status', '') AS state, w.status->'metadata' AS metadata
    FROM install_workflows w
    JOIN live_installs i ON w.owner_type = 'installs' AND w.owner_id = i.id AND w.org_id = i.org_id
    WHERE w.deleted_at = 0
), current_workflows AS (
    SELECT * FROM workflow_records
    WHERE (finished_at IS NULL AND state NOT IN ('success', 'cancelled', 'error', 'discarded', 'not-attempted', 'user-skipped', 'auto-skipped'))
        OR state IN ('approval-awaiting', 'failed-pending-retry')
        OR (state = 'error' AND metadata->>'awaiting_retry' = 'true'
            AND COALESCE(metadata->>'stopped', 'false') <> 'true'
            AND COALESCE(metadata->>'retries_exhausted', 'false') <> 'true')
), workflow_counts AS (
    SELECT %s,
        CASE WHEN type = '' OR type IS NULL THEN 'unset' WHEN type = ANY($1::text[]) THEN type ELSE 'other' END AS type,
        COALESCE(plan_only, false) AS plan_only,
        CASE WHEN state = '' THEN 'unset' WHEN state = ANY($2::text[]) THEN state ELSE 'other' END AS state,
        COUNT(*) AS value,
        CASE WHEN bool_and(COALESCE(created_at > 'epoch'::timestamptz AND created_at <= transaction_timestamp(), false))
            THEN EXTRACT(EPOCH FROM MIN(created_at))::double precision END AS oldest
    FROM current_workflows GROUP BY 1, 2, 3, 4, 5, 6
)
`

func newWorkflowFamily(meter metric.Meter, level string) (*family, error) {
	count, err := meter.Int64ObservableGauge("nuon.workflow.current.count", metric.WithUnit("{workflow}"),
		metric.WithDescription("Recorded current install-owned workflows by type, plan-only mode and status; not steps, attempts or live executions."))
	if err != nil {
		return nil, err
	}
	oldest, err := meter.Float64ObservableGauge("nuon.workflow.oldest_created_at", metric.WithUnit("s"),
		metric.WithDescription("Earliest creation time in the same current workflow bucket; omit empty or invalid timing. Aggregate with MIN, not SUM; not time in a wait."))
	if err != nil {
		return nil, err
	}
	types := make([]string, 0)
	for _, workflowType := range app.AllWorkflowTypes() {
		types = append(types, string(workflowType))
	}
	types = append(types, "unset", "other")
	projection := "org_id, app_id, install_id"
	groups := "SELECT org_id, app_id, id AS install_id FROM live_installs"
	if level == "summary" {
		projection = "org_id, app_id, ''::text AS install_id"
		groups = "SELECT org_id, id AS app_id, ''::text AS install_id FROM live_apps"
	} else if level == "fleet" {
		projection = "''::text AS org_id, ''::text AS app_id, ''::text AS install_id"
	}
	query := retainedResourcesSQL + fmt.Sprintf(currentWorkflowsSQL, projection)
	if level == "fleet" {
		query += "SELECT * FROM workflow_counts"
	} else {
		query += ", groups AS (" + groups + `)
SELECT g.org_id, g.app_id, g.install_id, COALESCE(c.type, 'unset'), COALESCE(c.plan_only, false),
    COALESCE(c.state, 'unset'), COALESCE(c.value, 0), c.oldest
FROM groups g LEFT JOIN workflow_counts c USING (org_id, app_id, install_id)
`
	}
	query += fmt.Sprintf(" LIMIT %d", maxSnapshotSeries+1)
	f := &family{
		name: "workflows", level: level, interval: time.Minute, ttl: 3 * time.Minute,
		metrics:     []string{"nuon.workflow.current.count", "nuon.workflow.oldest_created_at"},
		instruments: []metric.Observable{count, oldest},
	}
	f.read = func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
		type point struct {
			options metric.ObserveOption
			count   int64
			oldest  *float64
		}
		points := make([]point, 0)
		type bucket struct {
			kind, state string
			plan        bool
		}
		seen := make(map[bucket]bool)
		timestamps := 0
		rows, err := tx.Query(ctx, query, types, workflowStates)
		if err != nil {
			return snapshot{}, err
		}
		defer rows.Close()
		for rows.Next() {
			if len(points) == maxSnapshotSeries {
				return snapshot{}, fmt.Errorf("%w: workflow counts exceed %d points", errSeriesLimit, maxSnapshotSeries)
			}
			var orgID, appID, installID string
			var b bucket
			var p point
			if err := rows.Scan(&orgID, &appID, &installID, &b.kind, &b.plan, &b.state, &p.count, &p.oldest); err != nil {
				return snapshot{}, err
			}
			attrs := []attribute.KeyValue{attribute.String("nuon.workflow.type", b.kind), attribute.String("nuon.workflow.status", b.state), attribute.Bool("nuon.workflow.plan_only", b.plan)}
			for _, id := range []struct{ key, value string }{
				{"nuon.org.id", orgID}, {"nuon.app.id", appID}, {"nuon.install.id", installID},
			} {
				if id.value != "" {
					attrs = append(attrs, attribute.String(id.key, id.value))
				}
			}
			p.options = metric.WithAttributes(attrs...)
			points = append(points, p)
			seen[b] = true
			if p.oldest != nil {
				timestamps++
			}
		}
		if err := rows.Err(); err != nil {
			return snapshot{}, err
		}
		if level == "fleet" {
			for _, kind := range types {
				for _, state := range workflowStates {
					for _, plan := range []bool{false, true} {
						if !seen[bucket{kind, state, plan}] {
							points = append(points, point{options: metric.WithAttributes(attribute.String("nuon.workflow.type", kind), attribute.String("nuon.workflow.status", state), attribute.Bool("nuon.workflow.plan_only", plan))})
						}
					}
				}
			}
		}
		return snapshot{
			series: map[string]int{f.metrics[0]: len(points), f.metrics[1]: timestamps},
			observe: func(o metric.Observer) {
				for _, p := range points {
					o.ObserveInt64(count, p.count, p.options)
					if p.oldest != nil {
						o.ObserveFloat64(oldest, *p.oldest, p.options)
					}
				}
			},
		}, nil
	}
	return f, nil
}
