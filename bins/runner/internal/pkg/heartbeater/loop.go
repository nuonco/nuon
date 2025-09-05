package heartbeater

import (
	"context"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/version"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

const (
	heartBeatErrBackoff time.Duration = time.Second * 5
)

func (h *HeartBeater) writeHeartBeat(ctx context.Context) error {
	tags := metrics.ToTags(
		map[string]string{
			"process": h.process,
		},
	)

	aliveDur := time.Since(h.startTS)
	req := &models.ServiceCreateRunnerHeartBeatRequest{
		AliveTime: generics.ToPtr(int64(aliveDur)),
		Version:   version.Version,
		Process:   h.process,
	}

	if _, err := h.apiClient.CreateHeartBeat(ctx, req); err != nil {
		return err
	}

	h.mw.Incr("heart_beat.incr", tags)
	h.mw.Timing("heart_beat.alive_time", aliveDur, tags)
	return nil
}

func (h *HeartBeater) loop(ctx context.Context) {
	ticker := time.NewTicker(h.settings.HeartBeatTimeout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		h.l.Info("recording heart beat")

		if err := h.writeHeartBeat(ctx); err != nil {
			h.l.Error("unable to write heart beat", zap.Error(err))
		}
	}
}
