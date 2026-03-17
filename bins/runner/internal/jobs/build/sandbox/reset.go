package sandbox

import "context"

func (h *handler) Reset(ctx context.Context) {
	h.state = nil
}
