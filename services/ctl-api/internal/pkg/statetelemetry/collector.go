package statetelemetry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
)

const (
	refreshInterval = 30 * time.Second
	queryTimeout    = 5 * time.Second
	sessionLifetime = 10 * time.Minute
	leaderTTL       = 180 * time.Second
	lockNamespace   = 1853189998
	lockID          = 1
)

type collector struct {
	mu           sync.RWMutex
	leader       bool
	verifiedAt   time.Time
	families     []*family
	registration metric.Registration
	l            *zap.Logger

	leadershipAttempts metric.Int64Counter
	refreshAttempts    metric.Int64Counter
	refreshDuration    metric.Float64Histogram
}

func Register(lc fx.Lifecycle, cfg *internal.Config, tc *telemetry.Config, provider metric.MeterProvider, l *zap.Logger) error {
	if !cfg.StateTelemetryEnabled || tc.Endpoint == "" {
		return nil
	}
	for _, key := range []attribute.Key{"nuon.control_plane.id", "service.instance.id"} {
		if tc.Resource == nil {
			return fmt.Errorf("fleet metrics require resource attribute %s", key)
		}
		value, ok := tc.Resource.Set().Value(key)
		if !ok || value.Type() != attribute.STRING || strings.TrimSpace(value.AsString()) == "" {
			return fmt.Errorf("fleet metrics require nonempty resource attribute %s in OTEL_RESOURCE_ATTRIBUTES", key)
		}
	}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/statetelemetry")
	inventory, err := newInventoryFamily(meter)
	if err != nil {
		return err
	}
	apps, err := newScopedInventoryFamily(meter, "apps")
	if err != nil {
		return err
	}
	orgs, err := newScopedInventoryFamily(meter, "orgs")
	if err != nil {
		return err
	}
	c, err := newCollector(provider, l, inventory, apps, orgs)
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
				c.run(ctx, func(ctx context.Context) (*pgx.Conn, error) {
					return psql.NewPrimaryListenerConn(ctx, cfg)
				})
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			err := c.registration.Unregister()
			select {
			case <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
	return nil
}

func newCollector(provider metric.MeterProvider, l *zap.Logger, families ...*family) (*collector, error) {
	c := &collector{l: l, families: families}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/statetelemetry")
	leader, err := meter.Int64ObservableGauge("nuon.fleet.snapshot.leader",
		metric.WithDescription("1 while this process holds a recently verified snapshot leadership session, otherwise 0."))
	if err != nil {
		return nil, err
	}
	verified, err := meter.Float64ObservableGauge("nuon.fleet.snapshot.leader.last_verified", metric.WithUnit("s"),
		metric.WithDescription("Unix time of the last successful leadership-session check; absent before acquiring leadership."))
	if err != nil {
		return nil, err
	}
	active, err := meter.Int64ObservableGauge("nuon.fleet.snapshot.active",
		metric.WithDescription("1 when a family's snapshot is fresh and owned by the current leader, otherwise 0."))
	if err != nil {
		return nil, err
	}
	started, err := meter.Float64ObservableGauge("nuon.fleet.snapshot.started_at", metric.WithUnit("s"),
		metric.WithDescription("Unix read-start time of the last successfully published family snapshot; unchanged on failure."))
	if err != nil {
		return nil, err
	}
	collected, err := meter.Float64ObservableGauge("nuon.fleet.snapshot.collected_at", metric.WithUnit("s"),
		metric.WithDescription("Unix publication time of the last successful family snapshot; unchanged on failure."))
	if err != nil {
		return nil, err
	}
	observed, err := meter.Float64ObservableGauge("nuon.fleet.snapshot.observed_at", metric.WithUnit("s"),
		metric.WithDescription("Unix callback observation time for this family's export frame, not database refresh time."))
	if err != nil {
		return nil, err
	}
	c.leadershipAttempts, err = meter.Int64Counter("nuon.fleet.snapshot.leadership.attempts", metric.WithUnit("{attempt}"),
		metric.WithDescription("Leadership acquisition attempts by outcome: acquired, standby, or error."))
	if err != nil {
		return nil, err
	}
	c.refreshAttempts, err = meter.Int64Counter("nuon.fleet.snapshot.refresh.attempts", metric.WithUnit("{attempt}"),
		metric.WithDescription("Completed family refresh attempts by outcome: success, error, cancelled, or series_limit."))
	if err != nil {
		return nil, err
	}
	c.refreshDuration, err = meter.Float64Histogram("nuon.fleet.snapshot.refresh.duration", metric.WithUnit("s"),
		metric.WithDescription("Family refresh duration including transaction cleanup, by outcome."),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10))
	if err != nil {
		return nil, err
	}
	instruments := []metric.Observable{leader, verified, active, started, collected, observed}
	for _, f := range c.families {
		instruments = append(instruments, f.instruments...)
	}
	c.registration, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		c.mu.RLock()
		defer c.mu.RUnlock()
		now := time.Now()
		isLeader := c.leader && now.Sub(c.verifiedAt) <= leaderTTL
		var leaderValue int64
		if isLeader {
			leaderValue = 1
		}
		o.ObserveInt64(leader, leaderValue)
		if !c.verifiedAt.IsZero() {
			o.ObserveFloat64(verified, float64(c.verifiedAt.UnixNano())/1e9)
		}
		for _, f := range c.families {
			opts := metric.WithAttributes(attribute.String("nuon.fleet.snapshot.family", f.name))
			o.ObserveFloat64(observed, float64(now.UnixNano())/1e9, opts)
			var activeValue int64
			if isLeader && f.active && now.Sub(f.startedAt) <= f.ttl {
				activeValue = 1
			}
			o.ObserveInt64(active, activeValue, opts)
			if !f.collectedAt.IsZero() {
				o.ObserveFloat64(started, float64(f.startedAt.UnixNano())/1e9, opts)
				o.ObserveFloat64(collected, float64(f.collectedAt.UnixNano())/1e9, opts)
			}
			if activeValue == 1 && f.values.observe != nil {
				f.values.observe(o)
			}
		}
		return nil
	}, instruments...)
	return c, err
}

func (c *collector) run(ctx context.Context, connect func(context.Context) (*pgx.Conn, error)) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	var conn *pgx.Conn
	var openedAt time.Time
	defer func() { c.release(conn) }()
	for {
		if ctx.Err() != nil {
			return
		}
		if conn != nil && time.Since(openedAt) >= sessionLifetime {
			c.release(conn)
			conn = nil
		}
		queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
		var err error
		if conn == nil {
			conn, err = connect(queryCtx)
			if err == nil {
				var acquired bool
				acquired, err = acquire(queryCtx, conn)
				outcome := "error"
				if err == nil && acquired {
					outcome = "acquired"
					openedAt = time.Now()
					c.l.Info("acquired fleet snapshot leadership")
				} else if err == nil {
					outcome = "standby"
					c.release(conn)
					conn = nil
				}
				c.leadershipAttempts.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
			} else {
				c.leadershipAttempts.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "error")))
			}
		} else {
			err = conn.Ping(queryCtx)
		}
		cancel()
		if err != nil {
			if ctx.Err() == nil {
				c.l.Warn("unable to verify fleet snapshot leadership", zap.Error(err))
			}
			c.release(conn)
			conn = nil
		}
		if conn != nil {
			c.mu.Lock()
			c.leader, c.verifiedAt = true, time.Now()
			c.mu.Unlock()
			for _, f := range c.families {
				if err := c.refreshIfDue(ctx, conn, f); err != nil && ctx.Err() == nil {
					c.l.Warn("unable to refresh fleet snapshot", zap.String("family", f.name), zap.Error(err))
				}
				if conn.IsClosed() || ctx.Err() != nil {
					c.release(conn)
					conn = nil
					break
				}
			}
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

func (c *collector) release(conn *pgx.Conn) {
	c.mu.Lock()
	wasLeader := c.leader
	c.leader = false
	for _, f := range c.families {
		f.values = snapshot{}
		f.active = false
		f.lastAttempt = time.Time{}
	}
	c.mu.Unlock()
	if wasLeader {
		c.l.Info("released fleet snapshot leadership")
	}
	if conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = conn.Close(ctx)
	}
}
