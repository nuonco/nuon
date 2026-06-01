package workflowmanager

import (
	"math/rand"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	defaultHistoryMax    = 10000
	defaultCheckInterval = 3 * time.Minute
)

// Manager manages the lifecycle of a long-running workflow, providing
// continue-as-new checks, alive checks, and expiry checks in a single
// background goroutine. Callers read Stopped and Restarted to decide
// how to proceed in their main workflow.Await loop.
type Manager struct {
	// Stopped is set to true when the backing entity no longer exists
	// or the workflow has expired.
	Stopped bool

	// Restarted is set to true when a continue-as-new is needed
	// (history too large, hint requested, or error recovery).
	Restarted bool

	opts options
}

type options struct {
	historyMax    int
	checkInterval time.Duration
	aliveChecker  func(ctx workflow.Context) (bool, error)
	canHint       CANHintChecker
	expiryChecker func(ctx workflow.Context) (*time.Time, error)
	mw            tmetrics.Writer
	onStopped     func(ctx workflow.Context)
	deferRestart  func() bool
}

// Option configures a Manager.
type Option func(*options)

// WithHistoryMax sets the maximum workflow history length before triggering
// continue-as-new. Defaults to 10000.
func WithHistoryMax(n int) Option {
	return func(o *options) { o.historyMax = n }
}

// WithCheckInterval sets how often the background goroutine runs checks.
// Defaults to 3 minutes. Up to 50% jitter is added automatically.
func WithCheckInterval(d time.Duration) Option {
	return func(o *options) { o.checkInterval = d }
}

// WithAliveChecker provides a function that verifies the backing entity
// still exists. When it returns (false, nil), the manager sets Stopped=true.
func WithAliveChecker(fn func(ctx workflow.Context) (bool, error)) Option {
	return func(o *options) { o.aliveChecker = fn }
}

// WithCANHintChecker provides a checker for externally-requested
// continue-as-new hints (e.g., metadata flags in the database).
func WithCANHintChecker(c CANHintChecker) Option {
	return func(o *options) { o.canHint = c }
}

// WithExpiryChecker provides a function that returns when the workflow
// should terminate. If the returned time is in the past, the manager
// sets Stopped=true.
func WithExpiryChecker(fn func(ctx workflow.Context) (*time.Time, error)) Option {
	return func(o *options) { o.expiryChecker = fn }
}

// WithMetricsWriter sets the metrics writer for reporting workflow size gauges.
func WithMetricsWriter(mw tmetrics.Writer) Option {
	return func(o *options) { o.mw = mw }
}

// WithOnStopped provides a callback invoked when the manager sets Stopped=true.
// Useful for writing terminal status to DB before the workflow exits.
func WithOnStopped(fn func(ctx workflow.Context)) Option {
	return func(o *options) { o.onStopped = fn }
}

// WithDeferRestart holds off continue-as-new while fn returns true. Stop
// decisions are still honored. Continue-as-new abandons in-flight updates, so
// restarting mid-phase would orphan it and be misread as a crash.
func WithDeferRestart(fn func() bool) Option {
	return func(o *options) { o.deferRestart = fn }
}

// New creates a Manager with the given options.
func New(opts ...Option) *Manager {
	o := options{
		historyMax:    defaultHistoryMax,
		checkInterval: defaultCheckInterval,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &Manager{opts: o}
}

// CANResponse holds diagnostic info from a continue-as-new check.
type CANResponse struct {
	WorkflowType  string `json:"workflow_type"`
	Namespace     string `json:"namespace"`
	HistoryLength int    `json:"history_length"`
	HistoryMax    int    `json:"history_max"`
	HintRequested bool   `json:"hint_requested"`
	Restarting    bool   `json:"restarting"`
}

// Start begins the background lifecycle goroutine. It periodically runs
// CAN checks, alive checks, and expiry checks, setting Stopped or
// Restarted as appropriate. Returns immediately.
func (m *Manager) Start(ctx workflow.Context) {
	workflow.Go(ctx, func(gCtx workflow.Context) {
		m.run(gCtx)
	})
}

// RunCANCheck performs a single continue-as-new check on demand.
// Returns whether a restart should be triggered, along with diagnostics.
func (m *Manager) RunCANCheck(ctx workflow.Context) (bool, *CANResponse) {
	l, _ := log.WorkflowLogger(ctx)
	return m.checkCAN(ctx, l)
}

func (m *Manager) run(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	for {
		if m.Stopped || m.Restarted {
			return
		}

		// Add up to 50% jitter to the check interval.
		interval := m.opts.checkInterval
		jitterMax := int(interval.Seconds() / 2)
		if jitterMax < 1 {
			jitterMax = 1
		}
		interval += time.Duration(rand.Intn(jitterMax)) * time.Second
		if err := workflow.Sleep(ctx, interval); err != nil {
			return
		}

		if m.Stopped || m.Restarted {
			return
		}

		// Check 1: continue-as-new (history size + hint).
		restarting, _ := m.checkCAN(ctx, l)
		if restarting {
			if m.restartDeferred() {
				if l != nil {
					l.Info("continue-as-new needed but deferred until in-flight phase completes")
				}
			} else {
				m.Restarted = true
				return
			}
		}

		// Check 2: alive check.
		if m.opts.aliveChecker != nil {
			alive, err := m.opts.aliveChecker(ctx)
			if err != nil {
				if m.restartDeferred() {
					if l != nil {
						l.Warn("alive check failed during in-flight phase; deferring restart", zap.Error(err))
					}
					continue
				}
				if l != nil {
					l.Warn("alive check failed, restarting to recover", zap.Error(err))
				}
				m.Restarted = true
				return
			}
			if !alive {
				if l != nil {
					l.Warn("backing entity no longer exists, stopping")
				}
				m.Stopped = true
				if m.opts.onStopped != nil {
					m.opts.onStopped(ctx)
				}
				return
			}
		}

		// Check 3: expiry check.
		if m.opts.expiryChecker != nil {
			expiresAt, err := m.opts.expiryChecker(ctx)
			if err != nil {
				if l != nil {
					l.Warn("expiry check failed", zap.Error(err))
				}
				continue
			}
			if expiresAt != nil && workflow.Now(ctx).After(*expiresAt) {
				if l != nil {
					l.Warn("workflow expired, stopping")
				}
				m.Stopped = true
				if m.opts.onStopped != nil {
					m.opts.onStopped(ctx)
				}
				return
			}
		}
	}
}

func (m *Manager) restartDeferred() bool {
	return m.opts.deferRestart != nil && m.opts.deferRestart()
}

func (m *Manager) checkCAN(ctx workflow.Context, l *zap.Logger) (bool, *CANResponse) {
	info := workflow.GetInfo(ctx)
	historyLen := info.GetCurrentHistoryLength()

	// Emit workflow size metric.
	if m.opts.mw != nil {
		tags := metrics.ToTags(map[string]string{
			"namespace":     info.Namespace,
			"workflow_type": info.WorkflowType.Name,
			"is_can":        strconv.FormatBool(info.ContinuedExecutionRunID != ""),
		})
		m.opts.mw.Gauge(ctx, "workflow.workflow_size", float64(historyLen), tags...)
	}

	resp := &CANResponse{
		WorkflowType:  info.WorkflowType.Name,
		Namespace:     info.Namespace,
		HistoryLength: historyLen,
		HistoryMax:    m.opts.historyMax,
	}

	// Check 1: history length exceeds threshold.
	if historyLen > m.opts.historyMax {
		if l != nil {
			l.Info("history length exceeded threshold, triggering continue-as-new",
				zap.Int("history_length", historyLen),
				zap.Int("history_max", m.opts.historyMax))
		}
		resp.Restarting = true
		return true, resp
	}

	// Check 2: external hint.
	if m.opts.canHint != nil {
		requested, err := m.opts.canHint.CheckCANHint(ctx)
		if err != nil {
			if l != nil {
				l.Warn("CAN hint check failed, restarting to recover", zap.Error(err))
			}
			resp.Restarting = true
			return true, resp
		}
		resp.HintRequested = requested
		if requested {
			if l != nil {
				l.Info("continue-as-new requested via hint")
			}
			if clearErr := m.opts.canHint.ClearCANHint(ctx); clearErr != nil && l != nil {
				l.Warn("unable to clear CAN hint", zap.Error(clearErr))
			}
			resp.Restarting = true
			return true, resp
		}
	}

	return false, resp
}
