package jobloop

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type statusCoalescer struct {
	jobID       string
	executionID string
	write       writeJobExecutionStatusFn
	l           *zap.Logger

	mu       sync.Mutex
	pending  *coalescedStatus
	closed   bool
	wake     chan struct{}
	doneCh   chan struct{}
	stopOnce sync.Once
	stopCh   chan struct{}
}

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

		// why: Use a fresh context decoupled from the job context so a
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
