package statetelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var inventoryCategories = []struct {
	kind   string
	states []string
}{
	{"org", []string{"active", "error", "provisioning", "deleting", "deprovisioning", "deprovisioned", "unset", "other"}},
	{"app", []string{"active", "error", "provisioning", "deprovisioning", "delete_queued", "unset", "other"}},
	{"install", []string{"provisioning", "provisioned", "deprovisioning", "deprovisioned", "reprovisioning", "unset", "other"}},
	{"component", []string{"active", "error", "deprovisioning", "delete_queued", "unset", "other"}},
	{"install_component", []string{"deleted", "delete_failed", "queued", "active", "inactive", "error", "noop", "planning", "syncing", "executing", "unknown", "pending", "disabled", "unset", "other"}},
	{"runner", []string{"error", "active", "pending", "provisioning", "deprovisioning", "deprovisioned", "reprovisioning", "offline", "awaiting-install-stack-run", "awaiting-heartbeat", "disabled", "unknown", "unset", "other"}},
}

const inventoryResourcesSQL = `
WITH live_orgs AS (
    SELECT id, COALESCE(NULLIF(status_v2->>'status', ''), status, '') AS state
    FROM orgs WHERE deleted_at = 0
), live_apps AS (
    SELECT a.id, a.org_id, COALESCE(NULLIF(a.status_v2->>'status', ''), a.status, '') AS state
    FROM apps a JOIN live_orgs o ON o.id = a.org_id
    WHERE a.deleted_at = 0
), live_installs AS (
    SELECT i.id, i.org_id, i.app_id, COALESCE(i.lifecycle_phase->>'phase', '') AS state
    FROM installs i JOIN live_apps a ON a.id = i.app_id AND a.org_id = i.org_id
    WHERE i.deleted_at = 0
), live_components AS (
    SELECT c.id, c.org_id, c.app_id, COALESCE(NULLIF(c.status_v2->>'status', ''), c.status, '') AS state
    FROM components c JOIN live_apps a ON a.id = c.app_id AND a.org_id = c.org_id
    WHERE c.deleted_at = 0
), resources AS (
    SELECT 'org' AS kind, state, '' AS runner_type, id AS org_id, '' AS app_id FROM live_orgs
    UNION ALL
    SELECT 'app', state, '', org_id, id FROM live_apps
    UNION ALL
    SELECT 'install', state, '', org_id, app_id FROM live_installs
    UNION ALL
    SELECT 'component', state, '', org_id, app_id FROM live_components
    UNION ALL
    SELECT 'install_component', COALESCE(NULLIF(ic.status_v2->>'status', ''), ic.status, ''), '', ic.org_id, i.app_id
    FROM install_components ic
    JOIN live_installs i ON i.id = ic.install_id AND i.org_id = ic.org_id
    JOIN live_components c ON c.id = ic.component_id AND c.org_id = ic.org_id AND c.app_id = i.app_id
    WHERE ic.deleted_at = 0
    UNION ALL
    SELECT 'runner', COALESCE(NULLIF(r.status_v2->>'status', ''), r.status, ''), rg.type, r.org_id, ''
    FROM runners r
    JOIN live_orgs o ON o.id = r.org_id
    JOIN runner_groups rg ON rg.id = r.runner_group_id AND rg.org_id = r.org_id AND rg.deleted_at = 0
    WHERE r.deleted_at = 0 AND rg.type = 'org' AND rg.owner_type = 'orgs' AND rg.owner_id = o.id
    UNION ALL
    SELECT 'runner', COALESCE(NULLIF(r.status_v2->>'status', ''), r.status, ''), rg.type, r.org_id, i.app_id
    FROM runners r
    JOIN runner_groups rg ON rg.id = r.runner_group_id AND rg.org_id = r.org_id AND rg.deleted_at = 0
    JOIN live_installs i ON i.id = rg.owner_id AND i.org_id = rg.org_id
    WHERE r.deleted_at = 0 AND rg.type = 'install' AND rg.owner_type = 'installs'
), normalized_resources AS (
SELECT kind,
    CASE WHEN state = '' THEN 'unset'
         WHEN state = ANY(CASE kind
             WHEN 'org' THEN $1::text[]
             WHEN 'app' THEN $2::text[]
             WHEN 'install' THEN $3::text[]
             WHEN 'component' THEN $4::text[]
             WHEN 'install_component' THEN $5::text[]
             WHEN 'runner' THEN $6::text[] END) THEN state
         ELSE 'other' END AS state,
    runner_type, org_id, app_id
FROM resources
)
`

const inventorySQL = inventoryResourcesSQL + `
SELECT kind, state, runner_type, COUNT(*) FROM normalized_resources
GROUP BY 1, 2, 3
`

const appInventorySQL = inventoryResourcesSQL + `
SELECT a.org_id, a.id, COALESCE(r.kind, ''), COALESCE(r.state, ''), COUNT(r.kind)
FROM (SELECT id, org_id FROM live_apps ORDER BY id LIMIT $7) a
LEFT JOIN normalized_resources r ON r.app_id = a.id AND r.org_id = a.org_id
    AND r.kind IN ('install', 'component', 'install_component', 'runner')
    AND (r.kind <> 'runner' OR r.runner_type = 'install')
GROUP BY 1, 2, 3, 4
`

const orgInventorySQL = inventoryResourcesSQL + `
SELECT o.id, '', COALESCE(r.kind, ''), COALESCE(r.state, ''), COUNT(r.kind)
FROM (SELECT id FROM live_orgs ORDER BY id LIMIT $7) o
LEFT JOIN normalized_resources r ON r.org_id = o.id AND r.kind = 'runner' AND r.runner_type = 'org'
GROUP BY 1, 2, 3, 4
`

func newInventoryFamily(meter metric.Meter) (*family, error) {
	gauge, err := meter.Int64ObservableGauge("nuon.inventory.current", metric.WithUnit("{resource}"),
		metric.WithDescription("Retained resources with live, tenant-consistent parents, by kind and recorded lifecycle state; runners split by group type."))
	if err != nil {
		return nil, err
	}

	return &family{
		name:        "inventory",
		interval:    5 * time.Minute,
		ttl:         15 * time.Minute,
		instruments: []metric.Observable{gauge},
		read: func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
			if _, err := tx.Exec(ctx, "SET LOCAL row_security=off"); err != nil {
				return snapshot{}, err
			}
			type bucket struct{ kind, state, runnerType string }
			counts := make(map[bucket]int64)
			args := make([]any, len(inventoryCategories))
			for idx, category := range inventoryCategories {
				args[idx] = category.states
				types := []string{""}
				if category.kind == "runner" {
					types = []string{"org", "install"}
				}
				for _, state := range category.states {
					for _, runnerType := range types {
						counts[bucket{category.kind, state, runnerType}] = 0
					}
				}
			}
			rows, err := tx.Query(ctx, inventorySQL, args...)
			if err != nil {
				return snapshot{}, err
			}
			defer rows.Close()
			for rows.Next() {
				var key bucket
				var count int64
				if err := rows.Scan(&key.kind, &key.state, &key.runnerType, &count); err != nil {
					return snapshot{}, err
				}
				counts[key] = count
			}
			if err := rows.Err(); err != nil {
				return snapshot{}, err
			}
			return snapshot{
				series: map[string]int{"nuon.inventory.current": len(counts)},
				observe: func(o metric.Observer) {
					for key, count := range counts {
						attrs := []attribute.KeyValue{
							attribute.String("resource.kind", key.kind),
							attribute.String("resource.state", key.state),
						}
						if key.kind == "runner" {
							attrs = append(attrs, attribute.String("nuon.runner.group.type", key.runnerType))
						}
						o.ObserveInt64(gauge, count, metric.WithAttributes(attrs...))
					}
				},
			}, nil
		},
	}, nil
}

func newScopedInventoryFamily(meter metric.Meter, scope string) (*family, error) {
	categories := inventoryCategories[2:]
	query, runnerType := appInventorySQL, "install"
	if scope == "orgs" {
		categories = inventoryCategories[5:]
		query, runnerType = orgInventorySQL, "org"
	}
	gauges := make(map[string]metric.Int64ObservableGauge)
	names := make(map[string]string)
	instruments := make([]metric.Observable, 0, len(categories))
	maxStates := 0
	for _, category := range categories {
		name := "nuon.inventory." + category.kind + ".current"
		if category.kind == "runner" {
			name = "nuon.inventory." + runnerType + "_runner.current"
		}
		gauge, err := meter.Int64ObservableGauge(name, metric.WithUnit("{resource}"),
			metric.WithDescription("Retained resources with live, tenant-consistent parents, by "+scope+" and recorded lifecycle state."))
		if err != nil {
			return nil, err
		}
		gauges[category.kind], names[category.kind] = gauge, name
		instruments = append(instruments, gauge)
		if len(category.states) > maxStates {
			maxStates = len(category.states)
		}
	}
	groupLimit := maxSnapshotSeries / maxStates
	return &family{
		name: "inventory." + scope, interval: 5 * time.Minute, ttl: 15 * time.Minute,
		instruments: instruments,
		read: func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
			if _, err := tx.Exec(ctx, "SET LOCAL row_security=off"); err != nil {
				return snapshot{}, err
			}
			args := make([]any, 0, len(inventoryCategories)+1)
			for _, category := range inventoryCategories {
				args = append(args, category.states)
			}
			args = append(args, groupLimit+1)
			rows, err := tx.Query(ctx, query, args...)
			if err != nil {
				return snapshot{}, err
			}
			defer rows.Close()
			type bucket struct{ orgID, appID, kind, state string }
			counts := make(map[bucket]int64)
			groups := make(map[[2]string]struct{})
			series := make(map[string]int)
			for rows.Next() {
				var key bucket
				var count int64
				if err := rows.Scan(&key.orgID, &key.appID, &key.kind, &key.state, &count); err != nil {
					return snapshot{}, err
				}
				group := [2]string{key.orgID, key.appID}
				if _, ok := groups[group]; !ok {
					if len(groups) == groupLimit {
						return snapshot{}, fmt.Errorf("%w: inventory.%s exceeds %d groups", errSeriesLimit, scope, groupLimit)
					}
					groups[group] = struct{}{}
					for _, category := range categories {
						for _, state := range category.states {
							counts[bucket{key.orgID, key.appID, category.kind, state}] = 0
							series[names[category.kind]]++
						}
					}
				}
				if key.kind != "" {
					counts[key] = count
				}
			}
			if err := rows.Err(); err != nil {
				return snapshot{}, err
			}
			return snapshot{
				series: series,
				observe: func(o metric.Observer) {
					for key, count := range counts {
						attrs := []attribute.KeyValue{
							attribute.String("nuon.org.id", key.orgID), attribute.String("resource.state", key.state),
						}
						if scope == "apps" {
							attrs = append(attrs, attribute.String("nuon.app.id", key.appID))
						}
						if key.kind == "runner" {
							attrs = append(attrs, attribute.String("nuon.runner.group.type", runnerType))
						}
						o.ObserveInt64(gauges[key.kind], count, metric.WithAttributes(attrs...))
					}
				},
			}, nil
		},
	}, nil
}
