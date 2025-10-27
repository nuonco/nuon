package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetCHTableMetricsRequest struct{}

type GetCHPendingInsertsRequest struct{}

type GetCHPartsPerPartitionRequest struct{}

type GetCHPartStatisticsRequest struct {
	Database string
	Table    string
}

type PendingAsyncInsert struct {
	N       int `gorm:"column:n"`
	Queries int `gorm:"column:queries"`
}

type CHPartsPerPartition struct {
	Table             string
	PartsPerPartition int
	PartitionID       string
}

type CHRowPerParts struct {
	NumPartsCreated int     `gorm:"column:num_parts_created"`
	MaxRowsPerPart  int     `gorm:"column:max_rows_per_part"`
	MinRowsPerPart  int     `gorm:"column:min_rows_per_part"`
	AvgRowsPerPart  float64 `gorm:"column:avg_rows_per_part"`
}

type CHActivePartStats struct {
	Name  string `gorm:"column:name"`
	Level int    `gorm:"column:level"`
	Rows  int    `gorm:"column:rows"`
}

type GetCHActivePartStatsRequest struct {
	Database string
	Table    string
}

// @temporal-gen activity
func (a *Activities) GetCHTableMetrics(ctx context.Context, req GetCHTableMetricsRequest) ([]app.CHTableSize, error) {
	return a.getCHTableSizes(ctx, a.chDB)
}

// @temporal-gen activity
func (a *Activities) GetCHPendingInserts(ctx context.Context, req GetCHPendingInsertsRequest) ([]PendingAsyncInsert, error) {
	return a.getPendingAsyncInserts(ctx, a.chDB)
}

// @temporal-gen activity
func (a *Activities) GetCHPartsPerPartition(ctx context.Context, req GetCHPartsPerPartitionRequest) ([]CHPartsPerPartition, error) {
	return a.getCHPartsPerPartition(ctx, a.chDB)
}

// @temporal-gen activity
func (a *Activities) GetCHRowsPerPartStats(ctx context.Context, req GetCHPartStatisticsRequest) ([]CHRowPerParts, error) {
	return a.getCHRowsPerPart(ctx, a.chDB)
}

// @temporal-gen activity
func (a *Activities) GetCHActivePartStats(ctx context.Context, req GetCHActivePartStatsRequest) ([]CHActivePartStats, error) {
	return a.getCHActivePartStats(ctx, a.chDB)
}

func (a *Activities) getCHTableSizes(ctx context.Context, db *gorm.DB) ([]app.CHTableSize, error) {
	var tables []app.CHTableSize

	res := db.WithContext(ctx).
		Find(&tables)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get table sizes: %w", res.Error)
	}

	return tables, nil
}

func (a *Activities) getPendingAsyncInserts(ctx context.Context, db *gorm.DB) ([]PendingAsyncInsert, error) {
	var results []PendingAsyncInsert

	res := db.WithContext(ctx).
		Raw(`
SELECT
    DENSE_RANK() OVER (ORDER BY hostName() ASC) AS n,
    value AS queries
FROM clusterAllReplicas(default, system.metrics)
WHERE metric = 'PendingAsyncInsert'
ORDER BY n ASC
`).Scan(&results)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get pending async inserts: %w", res.Error)
	}

	return results, nil
}

func (a *Activities) getCHPartsPerPartition(ctx context.Context, db *gorm.DB) ([]CHPartsPerPartition, error) {
	query := `
		SELECT
			concat(database, '.', table) AS table,
			count() AS parts_per_partition,
			partition_id
		FROM clusterAllReplicas(default, system.parts)
		WHERE active AND (database != 'system')
		GROUP BY
			database,
			table,
			partition_id
		HAVING parts_per_partition > 1
		ORDER BY parts_per_partition DESC;
	`

	rows, err := db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var metrics []CHPartsPerPartition
	for rows.Next() {
		var metric CHPartsPerPartition
		if err := rows.Scan(&metric.Table, &metric.PartsPerPartition, &metric.PartitionID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return metrics, nil
}

func (a *Activities) getCHRowsPerPart(ctx context.Context, db *gorm.DB) ([]CHRowPerParts, error) {
	query := `
		SELECT
			count(*) AS num_parts_created,
			max(rows) AS max_rows_per_part,
			min(rows) AS min_rows_per_part,
			avg(rows) AS avg_rows_per_part
		FROM system.parts
		WHERE (active = 1) AND (database = ?) AND (table = ?)
	`

	var stats []CHRowPerParts
	rows, err := db.WithContext(ctx).Raw(query, "ctl_api", "runner_heart_beats").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var stat CHRowPerParts
		if err := rows.Scan(&stat.NumPartsCreated, &stat.MaxRowsPerPart, &stat.MinRowsPerPart, &stat.AvgRowsPerPart); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (a *Activities) getCHActivePartStats(ctx context.Context, db *gorm.DB) ([]CHActivePartStats, error) {
	query := `
		SELECT
			name,
			level,
			rows
		FROM system.parts
		WHERE (database = ?) AND (table = ?) AND active
		ORDER BY name ASC
	`

	var parts []CHActivePartStats
	res := db.WithContext(ctx).Raw(query, "ctl_api", "runner_heart_beats").Scan(&parts)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to execute query: %w", res.Error)
	}

	return parts, nil
}
