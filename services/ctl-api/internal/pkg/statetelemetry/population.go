package statetelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type populationMetric struct {
	name, unit, description string
	group, dimension        string
	states                  []string
	empty                   string
	source                  string
	detail, timestamp       bool
}

func (m populationMetric) query(level string) string {
	if m.detail {
		return retainedResourcesSQL + m.source + fmt.Sprintf(" LIMIT %d", maxSnapshotSeries+1)
	}
	group := m.group
	if level == "fleet" {
		group = "fleet"
	} else if level == "summary" && group == "install" {
		group = "app"
	}
	projection, groups := "''::text, ''::text, ''::text", ""
	switch group {
	case "org":
		projection, groups = "org_id, ''::text, ''::text", "SELECT id AS org_id, ''::text AS app_id, ''::text AS install_id FROM live_orgs"
	case "app":
		projection, groups = "org_id, app_id, ''::text", "SELECT org_id, id AS app_id, ''::text AS install_id FROM live_apps"
	case "install":
		projection, groups = "org_id, app_id, install_id", "SELECT org_id, app_id, id AS install_id FROM live_installs"
	}
	query := retainedResourcesSQL + ", population AS (" + m.source + `), normalized(org_id, app_id, install_id, state) AS (
SELECT ` + projection + `,
    CASE WHEN state = '' THEN 'unset' WHEN state = ANY($1::text[]) THEN state ELSE 'other' END
FROM population
), counts AS (
SELECT org_id, app_id, install_id, state, COUNT(*) AS value FROM normalized GROUP BY 1, 2, 3, 4
)
`
	if group == "fleet" {
		return query + "SELECT org_id, app_id, install_id, state, value FROM counts"
	}
	return query + ", groups AS (" + groups + `)
SELECT g.org_id, g.app_id, g.install_id, COALESCE(c.state, $2::text), COALESCE(c.value, 0)
FROM groups g LEFT JOIN counts c USING (org_id, app_id, install_id)
` + fmt.Sprintf("LIMIT %d", maxSnapshotSeries+1)
}

func newPopulationFamily(meter metric.Meter, level, name string, interval, ttl time.Duration, definitions ...populationMetric) (*family, error) {
	f := &family{name: name, level: level, interval: interval, ttl: ttl}
	readers := make([]func(context.Context, pgx.Tx) (snapshot, error), 0, len(definitions))
	for _, definition := range definitions {
		if definition.detail && level != "detailed" {
			continue
		}
		m := definition
		query := m.query(level)
		var gauge metric.Int64ObservableGauge
		var timestamp metric.Float64ObservableGauge
		var err error
		if m.timestamp {
			timestamp, err = meter.Float64ObservableGauge(m.name, metric.WithUnit(m.unit), metric.WithDescription(m.description))
			f.instruments = append(f.instruments, timestamp)
		} else {
			gauge, err = meter.Int64ObservableGauge(m.name, metric.WithUnit(m.unit), metric.WithDescription(m.description))
			f.instruments = append(f.instruments, gauge)
		}
		if err != nil {
			return nil, err
		}
		f.metrics = append(f.metrics, m.name)
		readers = append(readers, func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
			type point struct {
				options metric.ObserveOption
				count   int64
				time    float64
			}
			points := make([]point, 0)
			seen := make(map[string]bool)
			args := []any{m.states}
			if m.detail {
				args = nil
			} else if level != "fleet" {
				args = append(args, m.empty)
			}
			rows, err := tx.Query(ctx, query, args...)
			if err != nil {
				return snapshot{}, err
			}
			defer rows.Close()
			for rows.Next() {
				if len(points) == maxSnapshotSeries {
					return snapshot{}, fmt.Errorf("%w: %s exceeds %d points", errSeriesLimit, m.name, maxSnapshotSeries)
				}
				var orgID, appID, installID, state string
				var p point
				value := any(&p.count)
				if m.timestamp {
					value = &p.time
				}
				if err := rows.Scan(&orgID, &appID, &installID, &state, value); err != nil {
					return snapshot{}, err
				}
				attrs := make([]attribute.KeyValue, 0, 4)
				for _, id := range []struct{ key, value string }{
					{"nuon.org.id", orgID}, {"nuon.app.id", appID}, {"nuon.install.id", installID},
				} {
					if id.value != "" {
						attrs = append(attrs, attribute.String(id.key, id.value))
					}
				}
				if m.dimension != "" {
					if m.dimension == "nuon.install.sandbox_mode" {
						attrs = append(attrs, attribute.Bool(m.dimension, state == "true"))
					} else {
						attrs = append(attrs, attribute.String(m.dimension, state))
					}
				}
				p.options = metric.WithAttributes(attrs...)
				points = append(points, p)
				seen[state] = true
			}
			if err := rows.Err(); err != nil {
				return snapshot{}, err
			}
			if level == "fleet" {
				for _, state := range m.states {
					if !seen[state] {
						points = append(points, point{options: metric.WithAttributes(attribute.String(m.dimension, state))})
					}
				}
			}
			return snapshot{
				series: map[string]int{m.name: len(points)},
				observe: func(o metric.Observer) {
					for _, p := range points {
						if m.timestamp {
							o.ObserveFloat64(timestamp, p.time, p.options)
						} else {
							o.ObserveInt64(gauge, p.count, p.options)
						}
					}
				},
			}, nil
		})
	}
	f.read = func(ctx context.Context, tx pgx.Tx) (snapshot, error) {
		values := make([]snapshot, 0, len(readers))
		series := make(map[string]int, len(readers))
		for _, read := range readers {
			value, err := read(ctx, tx)
			if err != nil {
				return snapshot{}, err
			}
			values = append(values, value)
			for name, count := range value.series {
				series[name] = count
			}
		}
		return snapshot{series: series, observe: func(o metric.Observer) {
			for _, value := range values {
				value.observe(o)
			}
		}}, nil
	}
	return f, nil
}
