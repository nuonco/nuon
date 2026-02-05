package tests

import (
	"context"
	"sync"

	enumsv1 "go.temporal.io/api/enums/v1"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
)

// FakeEventLoopClient is a test implementation of eventloop.Client that records signals.
type FakeEventLoopClient struct {
	mu      sync.Mutex
	signals []CapturedSignal
}

// CapturedSignal holds information about a signal that was sent.
type CapturedSignal struct {
	ID     string
	Signal eventloop.Signal
}

// NewFakeEventLoopClient creates a new fake event loop client for testing.
func NewFakeEventLoopClient() *FakeEventLoopClient {
	return &FakeEventLoopClient{
		signals: make([]CapturedSignal, 0),
	}
}

// Send implements eventloop.Client by recording the signal.
func (f *FakeEventLoopClient) Send(ctx context.Context, id string, signal eventloop.Signal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.signals = append(f.signals, CapturedSignal{
		ID:     id,
		Signal: signal,
	})
}

// Cancel implements eventloop.Client (no-op for testing).
func (f *FakeEventLoopClient) Cancel(ctx context.Context, namespace, id string) error {
	return nil
}

// GetWorkflowStatus implements eventloop.Client (returns completed for testing).
func (f *FakeEventLoopClient) GetWorkflowStatus(ctx context.Context, namespace string, workflowID string) (enumsv1.WorkflowExecutionStatus, error) {
	return enumsv1.WORKFLOW_EXECUTION_STATUS_COMPLETED, nil
}

// GetWorkflowCount implements eventloop.Client (returns 1 for testing).
func (f *FakeEventLoopClient) GetWorkflowCount(ctx context.Context, namespace string, workflowID string) (int64, error) {
	return 1, nil
}

// GetSignals returns all captured signals.
func (f *FakeEventLoopClient) GetSignals() []CapturedSignal {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.signals
}

// Reset clears all captured signals.
func (f *FakeEventLoopClient) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.signals = make([]CapturedSignal, 0)
}

// Verify FakeEventLoopClient implements eventloop.Client
var _ eventloop.Client = (*FakeEventLoopClient)(nil)
