package helm

import "context"

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return h.state.outputs, nil
}
