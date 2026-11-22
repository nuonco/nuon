package workflow

import "context"

func (h *handler) Reset(ctx context.Context) error {
	h.state = &handlerState{}
	return nil
}
