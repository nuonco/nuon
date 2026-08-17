package cmd

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nuonco/nuon/pkg/runner/airgap"
)

func TestRunAirgapHeartbeatOverwritesUntilCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	writes := make(chan []byte, 4)
	done := make(chan struct{})
	startedAt := time.Now().Add(-time.Minute).UTC()

	go func() {
		defer close(done)
		runAirgapHeartbeat(ctx, airgap.RunnerHeartbeat{
			RunnerID: "runner-1", SessionID: "session-1", Version: "v1", BundleDigest: "sha256:bundle", StartedAt: startedAt,
		}, 5*time.Millisecond, func(_ context.Context, raw []byte) error {
			writes <- raw
			return nil
		}, zaptest.NewLogger(t))
	}()

	var first, second airgap.RunnerHeartbeat
	require.NoError(t, json.Unmarshal(<-writes, &first))
	require.NoError(t, json.Unmarshal(<-writes, &second))
	require.Equal(t, "runner-1", first.RunnerID)
	require.Equal(t, "session-1", first.SessionID)
	require.Equal(t, startedAt, first.StartedAt)
	require.True(t, second.ObservedAt.After(first.ObservedAt))

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("heartbeat writer did not stop after cancellation")
	}
}

func TestRunAirgapHeartbeatRetriesWriteErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	var mu sync.Mutex
	attempts := 0

	go func() {
		defer close(done)
		runAirgapHeartbeat(ctx, airgap.RunnerHeartbeat{}, 5*time.Millisecond, func(context.Context, []byte) error {
			mu.Lock()
			defer mu.Unlock()
			attempts++
			if attempts == 1 {
				return context.DeadlineExceeded
			}
			cancel()
			return nil
		}, zaptest.NewLogger(t))
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("heartbeat writer did not retry")
	}
	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, attempts, 2)
}
