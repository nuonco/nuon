package tests

import (
	"context"
	"sync"

	enumsv1 "go.temporal.io/api/enums/v1"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
)

// MockEventLoopClient is a test implementation of eventloop.Client that records signals.
type MockEventLoopClient struct {
	mu      sync.Mutex
	signals []CapturedSignal
}

// CapturedSignal holds information about a signal that was sent.
type CapturedSignal struct {
	ID     string
	Signal eventloop.Signal
}

// NewMockEventLoopClient creates a new mock event loop client for testing.
func NewMockEventLoopClient() *MockEventLoopClient {
	return &MockEventLoopClient{
		signals: make([]CapturedSignal, 0),
	}
}

// Send implements eventloop.Client by recording the signal.
func (f *MockEventLoopClient) Send(ctx context.Context, id string, signal eventloop.Signal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.signals = append(f.signals, CapturedSignal{
		ID:     id,
		Signal: signal,
	})
}

// Cancel implements eventloop.Client (no-op for testing).
func (f *MockEventLoopClient) Cancel(ctx context.Context, namespace, id string) error {
	return nil
}

// GetWorkflowStatus implements eventloop.Client (returns completed for testing).
func (f *MockEventLoopClient) GetWorkflowStatus(ctx context.Context, namespace string, workflowID string) (enumsv1.WorkflowExecutionStatus, error) {
	return enumsv1.WORKFLOW_EXECUTION_STATUS_COMPLETED, nil
}

// GetWorkflowCount implements eventloop.Client (returns 1 for testing).
func (f *MockEventLoopClient) GetWorkflowCount(ctx context.Context, namespace string, workflowID string) (int64, error) {
	return 1, nil
}

// GetSignals returns all captured signals.
func (f *MockEventLoopClient) GetSignals() []CapturedSignal {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.signals
}

// Reset clears all captured signals.
func (f *MockEventLoopClient) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.signals = make([]CapturedSignal, 0)
}

// Verify MockEventLoopClient implements eventloop.Client
var _ eventloop.Client = (*MockEventLoopClient)(nil)
