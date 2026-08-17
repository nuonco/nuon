package cmd

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap"
)

type airgapHeartbeatWriter func(context.Context, []byte) error

func runAirgapHeartbeat(ctx context.Context, heartbeat airgap.RunnerHeartbeat, interval time.Duration, write airgapHeartbeatWriter, logger *zap.Logger) {
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
