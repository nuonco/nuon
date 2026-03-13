package monitor

import (
	"context"
	"time"
)

const (
	runnerServiceCheckInterval = 15 * time.Second
	vmResourceCheckInterval    = 5 * time.Minute
)

func (h *Monitor) loop(ctx context.Context) {
	runnerTicker := time.NewTicker(runnerServiceCheckInterval)
	vmTicker := time.NewTicker(vmResourceCheckInterval)
	defer runnerTicker.Stop()
	defer vmTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-runnerTicker.C:
			h.settings.Refresh(ctx)
			h.checkRunnerService(ctx)
		case <-vmTicker.C:
			h.checkVMResources(ctx)
		}
	}
}
