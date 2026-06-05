package jobloop

import (
	"context"
	"sync"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
)

// statusCoalescer is a per-execution single-writer that drops intermediate
// non-terminal status pings while the previous write is still in flight.
// Each synchronous step-boundary update is a 50-200ms round-trip, and the
// dashboard only renders the latest non-terminal status, so dropping the
// intermediate values is safe and removes that latency from the init /
// finalize edges.
//
// Terminal statuses MUST land in order, so they go through WriteTerminal,
// which stops the background loop, drains it, then writes synchronously —
// guaranteeing a queued non-terminal update can never overwrite a final one.
type statusCoalescer struct {
	jobID       string
	executionID string
	write       writeJobExecutionStatusFn
	l           *zap.Logger

	// pending holds the latest queued non-terminal update; the mutex
	// guards coalesce-on-enqueue (replacing any earlier pending value).
	// wake is buffered to 1 so EnqueueNonTerminal never blocks the caller.
	mu       sync.Mutex
	pending  *coalescedStatus
	closed   bool
	wake     chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
	stopCh   chan struct{}
}

// writeJobExecutionStatusFn is the existing retry-wrapped update path
// from `exec_job_step.go`. Injecting it keeps coalescer logic free of the
// retry/metrics policy.
type writeJobExecutionStatusFn func(ctx context.Context, jobID, executionID string, status models.AppRunnerJobExecutionStatus, description string) error

type coalescedStatus struct {
	status      models.AppRunnerJobExecutionStatus
	description string
}

func newStatusCoalescer(jobID, executionID string, l *zap.Logger, write writeJobExecutionStatusFn) *statusCoalescer {
	c := &statusCoalescer{
		jobID:       jobID,
		executionID: executionID,
		write:       write,
		l:           l,
		wake:        make(chan struct{}, 1),
		doneCh:      make(chan struct{}),
		stopCh:      make(chan struct{}),
	}
	go c.run()
	return c
}

// EnqueueNonTerminal records `status` as the latest pending non-terminal
// update and signals the background writer. Returns immediately. Failures
// inside the writer are logged but never bubbled back; the contract is
// best-effort, because the next status update will overwrite it anyway.
func (c *statusCoalescer) EnqueueNonTerminal(status models.AppRunnerJobExecutionStatus, description string) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.pending = &coalescedStatus{status: status, description: description}
	c.mu.Unlock()

	select {
	case c.wake <- struct{}{}:
	default:
	}
}

// WriteTerminal stops the background writer, drops any in-flight pending
// non-terminal update, then writes the terminal status synchronously
// with the existing retry policy. Safe to call from panic recovery —
// idempotent via stopOnce so the deferred guard in `executeJob` can
// call it again without double-writing.
func (c *statusCoalescer) WriteTerminal(ctx context.Context, status models.AppRunnerJobExecutionStatus, description string) error {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.pending = nil
		c.mu.Unlock()
		close(c.stopCh)
		<-c.doneCh
	})
	return c.write(ctx, c.jobID, c.executionID, status, description)
}

// Close stops the background writer without writing a terminal status.
// Used as a `defer` guard so a panic before WriteTerminal still drains
// the goroutine. No-op after the first WriteTerminal / Close call.
func (c *statusCoalescer) Close() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.pending = nil
		c.mu.Unlock()
		close(c.stopCh)
		<-c.doneCh
	})
}

func (c *statusCoalescer) run() {
	defer close(c.doneCh)
	for {
		select {
		case <-c.stopCh:
			return
		case <-c.wake:
		}

		c.mu.Lock()
		next := c.pending
		c.pending = nil
		c.mu.Unlock()
		if next == nil {
			continue
		}

		// Use a fresh context decoupled from the job context so a
		// step that just returned doesn't cancel the trailing
		// status write. Bound it at 10s — the underlying write has
		// its own retry, this is a hard ceiling so we don't pile
		// up writes if the API is wedged.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := c.write(ctx, c.jobID, c.executionID, next.status, next.description); err != nil {
			c.l.Warn("coalesced status update failed",
				zap.String("status", string(next.status)),
				zap.Error(err),
			)
		}
		cancel()
	}
}

// isTerminalExecutionStatus identifies statuses that must be written
// synchronously and in order. Anything else is treated as a step-boundary
// transition the dashboard renders as a current-state string.
func isTerminalExecutionStatus(status models.AppRunnerJobExecutionStatus) bool {
	switch status {
	case models.AppRunnerJobExecutionStatusFinished,
		models.AppRunnerJobExecutionStatusFailed,
		models.AppRunnerJobExecutionStatusTimedDashOut,
		models.AppRunnerJobExecutionStatusCancelled:
		return true
	}
	return false
}
