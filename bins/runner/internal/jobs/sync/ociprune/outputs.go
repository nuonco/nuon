package ociprune

import (
	"context"
	"encoding/json"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	if h.state == nil || h.state.result == nil {
		return map[string]interface{}{}, nil
	}

	byts, err := json.Marshal(h.state.result)
	if err != nil {
		return nil, err
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(byts, &obj); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"prune": obj,
	}, nil
}
