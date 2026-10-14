package statetelemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
)

const maxSnapshotSeries = telemetry.ObservableGaugeCardinalityLimit - 1

var errSeriesLimit = errors.New("snapshot metric series limit exceeded")

type snapshot struct {
	observe func(metric.Observer)
	series  map[string]int
}

type family struct {
	name        string
	interval    time.Duration
	ttl         time.Duration
	instruments []metric.Observable
	read        func(context.Context, pgx.Tx) (snapshot, error)

	lastAttempt time.Time
	retryDelay  time.Duration
	startedAt   time.Time
	collectedAt time.Time
	active      bool
	values      snapshot
}

func (c *collector) refreshIfDue(ctx context.Context, conn *pgx.Conn, f *family) error {
	started := time.Now()
	interval := f.interval
	if f.retryDelay > 0 {
		interval = f.retryDelay
	}
	if started.Sub(f.lastAttempt) < interval {
		return nil
	}
	f.lastAttempt = started
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	values, err := readSnapshot(ctx, conn, f.read)
	if err == nil {
		c.mu.Lock()
		published := time.Now()
		err = ctx.Err()
		if err == nil && published.Sub(started) > f.ttl {
			err = fmt.Errorf("snapshot expired before publication")
		}
		if err == nil {
			f.values, f.startedAt, f.collectedAt, f.active = values, started, published, true
		}
		c.mu.Unlock()
	}
	completed := time.Now()
	var pgErr *pgconn.PgError
	if errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &pgErr) && (pgErr.Code == "57014" || pgErr.Code == "55P03")) {
		f.retryDelay = min(max(refreshInterval, f.retryDelay*2), f.interval)
	} else {
		f.retryDelay = 0
	}
	outcome := "success"
	if err != nil {
		outcome = "error"
		if errors.Is(err, errSeriesLimit) {
			outcome = "series_limit"
		} else if ctx.Err() != nil {
			outcome = "cancelled"
		}
	}
	opts := metric.WithAttributes(attribute.String("nuon.fleet.snapshot.family", f.name), attribute.String("outcome", outcome))
	c.refreshAttempts.Add(ctx, 1, opts)
	c.refreshDuration.Record(ctx, completed.Sub(started).Seconds(), opts)
	return err
}

func readSnapshot(ctx context.Context, conn *pgx.Conn, read func(context.Context, pgx.Tx) (snapshot, error)) (snapshot, error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return snapshot{}, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout='5s'; SET LOCAL lock_timeout='100ms';
SET LOCAL work_mem='4MB'; SET LOCAL max_parallel_workers_per_gather=0`); err != nil {
		return snapshot{}, err
	}
	values, err := read(ctx, tx)
	if err != nil {
		return snapshot{}, err
	}
	for name, count := range values.series {
		if count > maxSnapshotSeries {
			return snapshot{}, fmt.Errorf("%w: %s has %d series, limit %d", errSeriesLimit, name, count, maxSnapshotSeries)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return snapshot{}, err
	}
	return values, nil
}
