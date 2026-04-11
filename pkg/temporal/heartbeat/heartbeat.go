package heartbeat

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

// WithHeartbeat starts periodic heartbeating, runs fn, then stops heartbeating.
// Use this to wrap blocking activity work so Temporal can detect dead workers.
func WithHeartbeat[T any](ctx context.Context, interval time.Duration, fn func(context.Context) (T, error)) (T, error) {
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				activity.RecordHeartbeat(ctx, nil)
			}
		}
	}()
	return fn(ctx)
}
