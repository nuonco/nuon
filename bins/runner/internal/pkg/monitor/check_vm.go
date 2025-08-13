package monitor

import "context"

func (h *Monitor) checkVMResources(ctx context.Context) error {
	// TODO: implement some basic VM monitoring here
	h.l.Info("checking vm resources")
	return nil
}
