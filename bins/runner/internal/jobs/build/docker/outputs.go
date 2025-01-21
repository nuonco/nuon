package docker

import "context"

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"image": map[string]interface{}{
			"tag": h.state.resultTag,
		},
	}, nil
}
