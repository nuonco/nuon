package heartbeater

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/version"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
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
		ProcessID: h.processRegistrar.ProcessID(),
	}

	if _, err := h.apiClient.CreateHeartBeat(ctx, req); err != nil {
		return err
	}

	// ey: How does this work on the runner side? Do we have a ddog agent running?
	// does this add noise for the api side metric?
	h.mw.Incr("runner.heart_beat", tags)
	h.mw.Timing("runner.heart_beat.alive_time", aliveDur, tags)
	return nil
}

func (h *HeartBeater) loop(ctx context.Context) {
	// Smear initial heartbeat across the interval window so concurrent runners
	// don't sync up and pile up requests on the API every interval. After the
	// initial offset, ticks fire at exact intervals.
	if h.settings.HeartBeatTimeout > 0 {
		offset := rand.N(h.settings.HeartBeatTimeout)
		select {
		case <-ctx.Done():
			return
		case <-time.After(offset):
		}
	}

	ticker := time.NewTicker(h.settings.HeartBeatTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		h.l.Info("recording heart beat")

		// Bound each heartbeat write so a hung HTTP request (e.g. a stalled
		// HTTP/2 stream where http.DefaultTransport has no
		// ResponseHeaderTimeout) cannot park the loop forever and silently
		// stop heartbeats from reaching the API.
		writeCtx, cancel := context.WithTimeout(ctx, heartBeatErrBackoff)
		err := h.writeHeartBeat(writeCtx)
		cancel()
		if err != nil {
			h.l.Error("unable to write heart beat", zap.Error(err))
		}
	}
}
