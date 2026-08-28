package cmd

import (
	"context"
	"encoding/json"
	"time"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"go.uber.org/zap"
)

type customerManagedHeartbeatWriter func(context.Context, []byte) error

func runCustomerManagedHeartbeat(ctx context.Context, heartbeat customermanaged.RunnerHeartbeat, interval time.Duration, write customerManagedHeartbeatWriter, logger *zap.Logger) {
	writeHeartbeat := func() {
		heartbeat.ObservedAt = time.Now().UTC()
		raw, err := json.MarshalIndent(heartbeat, "", "  ")
		if err == nil {
			err = write(ctx, append(raw, '\n'))
		}
		if err != nil && ctx.Err() == nil {
			logger.Warn("publish runner heartbeat", zap.Error(err))
		}
	}

	writeHeartbeat()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			writeHeartbeat()
		}
	}
}
