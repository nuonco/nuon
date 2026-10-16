package statetelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const configurationSQL = `
, live_branches AS (
    SELECT b.id, b.org_id, b.app_id FROM app_branches b
    JOIN live_apps a ON a.id = b.app_id AND a.org_id = b.org_id WHERE b.deleted_at = 0
), branch_configs AS (
    SELECT b.org_id, b.app_id, ''::text AS install_id, b.id AS branch_id, c.*
    FROM live_branches b JOIN LATERAL (
        SELECT c.id AS config_id, c.created_at, c.status, c.status_v2, c.labels
        FROM app_configs c WHERE c.app_id = b.app_id AND c.org_id = b.org_id
            AND c.app_branch_id = b.id AND c.deleted_at = 0
        ORDER BY c.created_at DESC LIMIT 1
    ) c ON true
    WHERE c.created_at IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM app_configs peer WHERE peer.app_id = b.app_id AND peer.org_id = b.org_id
            AND peer.app_branch_id = b.id AND peer.deleted_at = 0
            AND peer.created_at = c.created_at AND peer.id != c.config_id
    )
    LIMIT %[1]d
), selected_configs AS (
    SELECT i.org_id, i.app_id, i.id AS install_id, COALESCE(b.id, '') AS branch_id,
        c.id AS config_id, c.created_at, c.status, c.status_v2, c.labels
    FROM live_installs i JOIN installs record ON record.id = i.id
    JOIN app_configs c ON c.id = record.app_config_id AND c.org_id = i.org_id AND c.app_id = i.app_id AND c.deleted_at = 0
    LEFT JOIN live_branches b ON b.id = c.app_branch_id AND b.org_id = i.org_id AND b.app_id = i.app_id
    WHERE %[2]t AND (COALESCE(c.app_branch_id, '') = '' OR b.id IS NOT NULL)
    LIMIT %[1]d
), candidates AS MATERIALIZED (
    SELECT false AS selected, * FROM branch_configs UNION ALL SELECT true, * FROM selected_configs
), ordinals AS MATERIALIZED (
    SELECT ranked.config_id, ranked.ordinal FROM (
        SELECT history.id AS config_id,
            CASE WHEN history.created_at IS NOT NULL AND COUNT(*) OVER (PARTITION BY history.app_id, history.created_at) = 1
                THEN RANK() OVER (PARTITION BY history.app_id ORDER BY history.created_at) END AS ordinal
        FROM app_configs history JOIN (SELECT DISTINCT app_id FROM candidates) apps ON apps.app_id = history.app_id
    ) ranked JOIN (SELECT DISTINCT config_id FROM candidates) needed ON needed.config_id = ranked.config_id
)
SELECT c.selected, c.org_id, c.app_id, c.install_id, c.branch_id,
    CASE WHEN COALESCE(NULLIF(c.status_v2->>'status', ''), c.status, '') = '' THEN 'unset'
         WHEN COALESCE(NULLIF(c.status_v2->>'status', ''), c.status) = ANY($1::text[])
            THEN COALESCE(NULLIF(c.status_v2->>'status', ''), c.status) ELSE 'other' END,
    CASE WHEN COALESCE(c.labels->>'source', '') = '' THEN 'unset'
         WHEN c.labels->>'source' = ANY($2::text[]) THEN c.labels->>'source' ELSE 'other' END,
    o.ordinal, CASE WHEN c.created_at > 'epoch'::timestamptz AND c.created_at <= transaction_timestamp()
        THEN EXTRACT(EPOCH FROM c.created_at)::double precision END
FROM candidates c JOIN ordinals o ON o.config_id = c.config_id
`

func newConfigurationFamily(meter metric.Meter, level string) (*family, error) {
	f := &family{name: "configurations", level: level, interval: time.Minute, ttl: 3 * time.Minute}
	ordinal, err := meter.Int64ObservableGauge("nuon.app_branch.config.latest_created.ordinal",
		metric.WithDescription("App-wide ordinal of the newest retained branch configuration; not the latest deployable release. Ambiguous ordinals are absent."))
	if err != nil {
		return nil, err
	}
	created, err := meter.Float64ObservableGauge("nuon.app_branch.config.latest_created.created_at", metric.WithUnit("s"),
		metric.WithDescription("Valid creation time of the newest retained branch configuration in Unix seconds, including preview/incomplete configurations; not deployability."))
	if err != nil {
		return nil, err
	}
	f.metrics = []string{"nuon.app_branch.config.latest_created.ordinal", "nuon.app_branch.config.latest_created.created_at"}
	f.instruments = []metric.Observable{ordinal, created}
	var selected metric.Int64ObservableGauge
	if level == "detailed" {
		selected, err = meter.Int64ObservableGauge("nuon.install.config.selected.ordinal",
			metric.WithDescription("App-wide ordinal of the install's selected configuration, not necessarily applied; branch identity is configuration provenance, not the tracked branch."))
		if err != nil {
			return nil, err
		}
		f.metrics = append(f.metrics, "nuon.install.config.selected.ordinal")
		f.instruments = append(f.instruments, selected)
	}
	query := retainedResourcesSQL + fmt.Sprintf(configurationSQL, maxSnapshotSeries+1, level == "detailed")
	f.read = func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
		type point struct {
			options  metric.ObserveOption
			selected bool
			ordinal  *int64
			created  *float64
		}
		points := make([]point, 0)
		series := make(map[string]int, len(f.metrics))
		for _, name := range f.metrics {
			series[name] = 0
		}
		rows, err := tx.Query(ctx, query, []string{"active", "pending", "syncing", "error", "outdated"}, []string{"git-run", "manual-run", "git-preview-run"})
		if err != nil {
			return snapshot{}, err
		}
		defer rows.Close()
		branches, installs := 0, 0
		for rows.Next() {
			var p point
			var orgID, appID, installID, branchID, status, source string
			if err := rows.Scan(&p.selected, &orgID, &appID, &installID, &branchID, &status, &source, &p.ordinal, &p.created); err != nil {
				return snapshot{}, err
			}
			if p.selected {
				installs++
			} else {
				branches++
			}
			if branches > maxSnapshotSeries || installs > maxSnapshotSeries {
				return snapshot{}, fmt.Errorf("%w: configuration candidates exceed %d per scope", errSeriesLimit, maxSnapshotSeries)
			}
			attrs := []attribute.KeyValue{attribute.String("nuon.org.id", orgID), attribute.String("nuon.app.id", appID),
				attribute.String("nuon.config.status", status), attribute.String("nuon.config.source", source)}
			if installID != "" {
				attrs = append(attrs, attribute.String("nuon.install.id", installID))
			}
			if branchID != "" {
				attrs = append(attrs, attribute.String("nuon.app_branch.id", branchID))
			}
			p.options = metric.WithAttributes(attrs...)
			if p.selected {
				p.created = nil
				if p.ordinal != nil {
					series[f.metrics[2]]++
				}
			} else {
				if p.ordinal != nil {
					series[f.metrics[0]]++
				}
				if p.created != nil {
					series[f.metrics[1]]++
				}
			}
			points = append(points, p)
		}
		if err := rows.Err(); err != nil {
			return snapshot{}, err
		}
		return snapshot{series: series, observe: func(o metric.Observer) {
			for _, p := range points {
				if p.ordinal != nil {
					gauge := ordinal
					if p.selected {
						gauge = selected
					}
					o.ObserveInt64(gauge, *p.ordinal, p.options)
				}
				if p.created != nil {
					o.ObserveFloat64(created, *p.created, p.options)
				}
			}
		}}, nil
	}
	return f, nil
}
