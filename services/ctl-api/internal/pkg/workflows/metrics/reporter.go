package workflowmetrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	refreshInterval = 30 * time.Second
	queryTimeout    = 5 * time.Second
	snapshotTTL     = 180 * time.Second
	// The two-key advisory lock is scoped to the primary database and this reporter.
	lockNamespace = 1853189998
	lockID        = 1
)

type reporter struct {
	mu           sync.RWMutex
	values       map[bucket]value
	collectedAt  time.Time
	active       bool
	leader       bool
	registration metric.Registration
	l            *zap.Logger
}

func Start(lc fx.Lifecycle, cfg *internal.Config, tc *telemetry.Config, provider metric.MeterProvider, l *zap.Logger) error {
	if tc.Endpoint == "" {
		return nil
	}
	r, err := newReporter(provider, l)
	if err != nil {
		return err
	}
	var cancel context.CancelFunc
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go func() {
				defer close(done)
				r.run(ctx, func(ctx context.Context) (*pgx.Conn, error) {
					return psql.NewPrimaryListenerConn(ctx, cfg)
				})
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			select {
			case <-done:
				return r.registration.Unregister()
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
	return nil
}

func newReporter(provider metric.MeterProvider, l *zap.Logger) (*reporter, error) {
	r := &reporter{l: l}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/workflows")
	current, err := meter.Int64ObservableGauge("nuon.workflow.current", metric.WithUnit("{workflow}"),
		metric.WithDescription("Current install provisioning and deployment workflows by persisted state."))
	if err != nil {
		return nil, err
	}
	oldest, err := meter.Float64ObservableGauge("nuon.workflow.oldest_created_at", metric.WithUnit("s"),
		metric.WithDescription("Unix creation time of the oldest workflow in this state, not the state entry time."))
	if err != nil {
		return nil, err
	}
	collected, err := meter.Float64ObservableGauge("nuon.workflow.snapshot.collected_at", metric.WithUnit("s"),
		metric.WithDescription("Unix time of the last successful workflow snapshot; unchanged on collection failure."))
	if err != nil {
		return nil, err
	}
	r.registration, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		r.mu.RLock()
		defer r.mu.RUnlock()
		if r.collectedAt.IsZero() {
			return nil
		}
		o.ObserveFloat64(collected, float64(r.collectedAt.UnixNano())/1e9)
		if !r.active || time.Since(r.collectedAt) > snapshotTTL {
			return nil
		}
		for _, typ := range workflowTypes {
			for _, state := range workflowStates {
				v := r.values[bucket{typ, state}]
				opts := metric.WithAttributes(attribute.String("workflow.type", typ), attribute.String("workflow.state", state))
				o.ObserveInt64(current, v.count, opts)
				if v.count > 0 {
					o.ObserveFloat64(oldest, float64(v.oldest.UnixNano())/1e9, opts)
				}
			}
		}
		return nil
	}, current, oldest, collected)
	return r, err
}

func (r *reporter) run(ctx context.Context, connect func(context.Context) (*pgx.Conn, error)) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	var conn *pgx.Conn
	var openedAt time.Time
	defer func() { r.release(conn) }()
	for {
		// Rotate before collecting so a healthy session does not lose an export cycle.
		if conn != nil && time.Since(openedAt) >= 10*time.Minute {
			r.release(conn)
			conn = nil
		}
		queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
		var err error
		if conn == nil {
			conn, err = connect(queryCtx)
			if err == nil {
				var acquired bool
				acquired, err = acquire(queryCtx, conn)
				if err == nil && !acquired {
					r.release(conn)
					conn = nil
				}
				if err == nil && acquired {
					r.mu.Lock()
					r.leader = true
					r.mu.Unlock()
					r.l.Info("acquired workflow metrics leadership")
				}
				openedAt = time.Now()
			}
		}
		if err == nil && conn != nil {
			err = r.refresh(queryCtx, conn)
		}
		cancel()
		if err != nil {
			if ctx.Err() == nil {
				r.l.Warn("unable to collect workflow metrics", zap.Error(err))
			}
			r.release(conn)
			conn = nil
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func acquire(ctx context.Context, conn *pgx.Conn) (bool, error) {
	var acquired bool
	err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1, $2)", lockNamespace, lockID).Scan(&acquired)
	return acquired, err
}

func (r *reporter) refresh(ctx context.Context, conn *pgx.Conn) error {
	// Timestamp the query start so slow queries cannot make old data appear newer.
	started := time.Now()
	values, err := readSnapshot(ctx, conn)
	if err != nil {
		return fmt.Errorf("read workflow snapshot: %w", err)
	}
	r.mu.Lock()
	r.values, r.collectedAt, r.active = values, started, true
	r.mu.Unlock()
	return nil
}

func (r *reporter) release(conn *pgx.Conn) {
	r.mu.Lock()
	r.active = false
	wasLeader := r.leader
	r.leader = false
	r.mu.Unlock()
	if wasLeader {
		r.l.Info("released workflow metrics leadership")
	}
	if conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = conn.Close(ctx)
	}
}
