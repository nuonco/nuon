package monitor

import (
	"context"
	"time"
)

func (h *Monitor) loop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 15)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		h.settings.Refresh(ctx)
		h.checkRunnerService(ctx)
		h.checkVMResources(ctx)
	}
}
