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

var deploymentStates = []string{"pending", "queued", "planning", "in-progress", "retrying", "approval-awaiting", "failed-pending-retry", "error", "success", "cancelled", "discarded", "not-attempted", "user-skipped", "auto-skipped", "approved", "unset", "other"}

const latestDeploymentsSQL = `
, latest_deployments AS (
    SELECT i.org_id, i.app_id, i.id AS install_id, COALESCE(w.type, 'none') AS type,
        CASE WHEN w.id IS NULL THEN 'never_deployed'
             WHEN COALESCE(w.status->>'status', '') = '' THEN 'unset'
             WHEN w.status->>'status' = ANY($2::text[]) THEN w.status->>'status'
             ELSE 'other' END AS state,
        CASE WHEN w.created_at > 'epoch'::timestamptz AND w.created_at <= transaction_timestamp()
            THEN EXTRACT(EPOCH FROM w.created_at)::double precision END AS created_at,
        CASE WHEN w.finished_at > 'epoch'::timestamptz AND w.finished_at <= transaction_timestamp()
            THEN EXTRACT(EPOCH FROM w.finished_at)::double precision END AS finished_at
    FROM live_installs i LEFT JOIN LATERAL (
        SELECT w.id, w.type, w.status, w.created_at, w.finished_at
        FROM install_workflows w
        WHERE w.owner_type = 'installs' AND w.owner_id = i.id AND w.org_id = i.org_id
            AND w.deleted_at = 0 AND w.plan_only = false AND w.type = ANY($1::text[])
        ORDER BY w.created_at DESC, w.id DESC LIMIT 1
    ) w ON true
), deployment_counts AS (
    SELECT %s, type, state, COUNT(*) AS value FROM latest_deployments GROUP BY 1, 2, 3, 4, 5
)
`

func newDeploymentFamily(meter metric.Meter, level string) (*family, error) {
	count, err := meter.Int64ObservableGauge("nuon.install.deploy.latest.count", metric.WithUnit("{install}"),
		metric.WithDescription("Installs by the newest non-plan deployment workflow's recorded type/status, including never_deployed; not successful rollouts or completion totals."))
	if err != nil {
		return nil, err
	}
	f := &family{
		name: "deployments", level: level, interval: time.Minute, ttl: 3 * time.Minute,
		metrics: []string{"nuon.install.deploy.latest.count"}, instruments: []metric.Observable{count},
	}
	var created, finished metric.Float64ObservableGauge
	if level == "detailed" {
		created, err = meter.Float64ObservableGauge("nuon.install.deploy.latest.created_at", metric.WithUnit("s"),
			metric.WithDescription("Valid recorded creation time of the selected latest deployment workflow in Unix seconds; absent without eligible history or valid timing."))
		if err != nil {
			return nil, err
		}
		finished, err = meter.Float64ObservableGauge("nuon.install.deploy.latest.finished_at", metric.WithUnit("s"),
			metric.WithDescription("Valid recorded finish time of that same latest deployment workflow in Unix seconds; not proof of success or terminality."))
		if err != nil {
			return nil, err
		}
		f.metrics = append(f.metrics, "nuon.install.deploy.latest.created_at", "nuon.install.deploy.latest.finished_at")
		f.instruments = append(f.instruments, created, finished)
	}
	types := make([]string, 0)
	for _, kind := range app.AllWorkflowTypes() {
		if kind.DeploymentType() != "" {
			types = append(types, string(kind))
		}
	}
	projection := "org_id, app_id, install_id"
	if level == "summary" {
		projection = "org_id, app_id, ''::text AS install_id"
	} else if level == "fleet" {
		projection = "''::text AS org_id, ''::text AS app_id, ''::text AS install_id"
	}
	query := retainedResourcesSQL + fmt.Sprintf(latestDeploymentsSQL, projection)
	switch level {
	case "detailed":
		query += "SELECT org_id, app_id, install_id, type, state, 1::bigint, created_at, finished_at FROM latest_deployments"
	case "summary":
		query += `
SELECT a.org_id, a.id, ''::text, COALESCE(c.type, 'none'), COALESCE(c.state, 'never_deployed'),
    COALESCE(c.value, 0), NULL::double precision, NULL::double precision
FROM live_apps a LEFT JOIN deployment_counts c ON c.org_id = a.org_id AND c.app_id = a.id`
	default:
		query += "SELECT org_id, app_id, install_id, type, state, value, NULL::double precision, NULL::double precision FROM deployment_counts"
	}
	query += fmt.Sprintf(" LIMIT %d", maxSnapshotSeries+1)
	f.read = func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
		type point struct {
			options           metric.ObserveOption
			count             int64
			created, finished *float64
		}
		type bucket struct{ kind, state string }
		points := make([]point, 0)
		seen := make(map[bucket]bool)
		series := map[string]int{f.metrics[0]: 0}
		if level == "detailed" {
			series[f.metrics[1]], series[f.metrics[2]] = 0, 0
		}
		rows, err := tx.Query(ctx, query, types, deploymentStates)
		if err != nil {
			return snapshot{}, err
		}
		defer rows.Close()
		for rows.Next() {
			if len(points) == maxSnapshotSeries {
				return snapshot{}, fmt.Errorf("%w: latest deployments exceed %d points", errSeriesLimit, maxSnapshotSeries)
			}
			var orgID, appID, installID string
			var b bucket
			var p point
			if err := rows.Scan(&orgID, &appID, &installID, &b.kind, &b.state, &p.count, &p.created, &p.finished); err != nil {
				return snapshot{}, err
			}
			attrs := []attribute.KeyValue{attribute.String("nuon.workflow.type", b.kind), attribute.String("nuon.workflow.status", b.state)}
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
			if p.created != nil {
				series[f.metrics[1]]++
			}
			if p.finished != nil {
				series[f.metrics[2]]++
			}
		}
		if err := rows.Err(); err != nil {
			return snapshot{}, err
		}
		if level == "fleet" {
			for _, kind := range types {
				for _, state := range deploymentStates {
					if !seen[bucket{kind, state}] {
						points = append(points, point{options: metric.WithAttributes(attribute.String("nuon.workflow.type", kind), attribute.String("nuon.workflow.status", state))})
					}
				}
			}
			if !seen[bucket{"none", "never_deployed"}] {
				points = append(points, point{options: metric.WithAttributes(attribute.String("nuon.workflow.type", "none"), attribute.String("nuon.workflow.status", "never_deployed"))})
			}
		}
		series[f.metrics[0]] = len(points)
		return snapshot{series: series, observe: func(o metric.Observer) {
			for _, p := range points {
				o.ObserveInt64(count, p.count, p.options)
				if p.created != nil {
					o.ObserveFloat64(created, *p.created, p.options)
				}
				if p.finished != nil {
					o.ObserveFloat64(finished, *p.finished, p.options)
				}
			}
		}}, nil
	}
	return f, nil
}
