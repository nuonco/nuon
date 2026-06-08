package service

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
)

const (
	runnerJobNotifyChannel = "runner_job_available_v1"

	// listenerSessionMaxAge tears down and reopens the listener connection well
	// inside the pool's 5m MaxConnLifetime and the ~15m IAM token window, so the
	// listener never runs on an aged or stale-auth connection. Reopening re-runs
	// the IAM token fetch in psql.NewPrimaryListenerConn.
	listenerSessionMaxAge = 4*time.Minute + 30*time.Second
	listenerMinReconnect  = 250 * time.Millisecond
	listenerMaxReconnect  = 10 * time.Second
)

type runnerJobNotifyPayload struct {
	RunnerID string `json:"runner_id"`
	JobID    string `json:"job_id"`
	Group    string `json:"group"`
}

// RunnerJobNotifyListener holds one primary-DB LISTEN connection per pod and
// fans NOTIFY events out to parked TailRunnerJobs handlers via the wake
// registry. It is intentionally best-effort: the long-poll handler keeps a poll
// backstop, so a dropped notify (reconnect, pod restart, RDS failover) only
// costs latency, never correctness.
type RunnerJobNotifyListener struct {
	cfg      *internal.Config
	l        *zap.Logger
	mw       metrics.Writer
	registry *RunnerJobWakeRegistry

	cancel context.CancelFunc
	done   chan struct{}
}

type RunnerJobNotifyListenerParams struct {
	fx.In

	Cfg      *internal.Config
	L        *zap.Logger
	MW       metrics.Writer
	Registry *RunnerJobWakeRegistry
}

// StartRunnerJobNotifyListener constructs the listener and binds its lifecycle
// to fx. Wired via fx.Invoke so the listener starts eagerly without anything
// having to depend on it.
func StartRunnerJobNotifyListener(p RunnerJobNotifyListenerParams, lc fx.Lifecycle) *RunnerJobNotifyListener {
	rl := &RunnerJobNotifyListener{
		cfg:      p.Cfg,
		l:        p.L.With(zap.String("component", "runner_job_notify_listener")),
		mw:       p.MW,
		registry: p.Registry,
		done:     make(chan struct{}),
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ctx, cancel := context.WithCancel(context.Background())
			rl.cancel = cancel
			go rl.run(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if rl.cancel != nil {
				rl.cancel()
			}
			select {
			case <-rl.done:
			case <-ctx.Done():
			}
			return nil
		},
	})

	return rl
}

func (rl *RunnerJobNotifyListener) run(ctx context.Context) {
	defer close(rl.done)

	backoff := listenerMinReconnect
	for ctx.Err() == nil {
		clean, err := rl.listenOnce(ctx)
		if ctx.Err() != nil {
			return
		}

		if clean {
			// Scheduled session rotation (age-out), not a fault: reset backoff
			// and reconnect promptly.
			backoff = listenerMinReconnect
		} else {
			rl.l.Warn("notify listener disconnected, reconnecting", zap.Error(err))
			rl.mw.Count("runner_job_tail.listener_error", 1, nil)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff + jitter(backoff)):
		}

		if !clean {
			backoff *= 2
			if backoff > listenerMaxReconnect {
				backoff = listenerMaxReconnect
			}
		}
	}
}

// listenOnce opens a connection, LISTENs, and pumps notifications until the
// session ages out or the connection drops. It returns clean=true when the exit
// was our own scheduled age-out (parent ctx still live), so the caller can treat
// it as a routine rotation rather than an error.
func (rl *RunnerJobNotifyListener) listenOnce(parent context.Context) (clean bool, err error) {
	ctx, cancel := context.WithTimeout(parent, listenerSessionMaxAge+jitter(30*time.Second))
	defer cancel()

	conn, err := psql.NewPrimaryListenerConn(ctx, rl.cfg)
	if err != nil {
		return false, err
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(ctx, "LISTEN "+runnerJobNotifyChannel); err != nil {
		return false, err
	}

	rl.mw.Gauge("runner_job_tail.listener_up", 1, nil)
	defer rl.mw.Gauge("runner_job_tail.listener_up", 0, nil)
	rl.l.Info("notify listener connected")

	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			// Our own session-age deadline elapsed while the service is still
			// running → routine rotation, not a fault.
			if parent.Err() == nil && ctx.Err() != nil {
				return true, nil
			}
			return false, err
		}

		var payload runnerJobNotifyPayload
		if jerr := json.Unmarshal([]byte(n.Payload), &payload); jerr != nil || payload.RunnerID == "" {
			rl.mw.Count("runner_job_tail.notify_payload_invalid", 1, nil)
			continue
		}

		rl.registry.Wake(payload.RunnerID)
	}
}
