package pulumi

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	if h.state == nil || h.state.outputs == nil {
		return map[string]interface{}{}, nil
	}

	// Round-trip through JSON to normalize types (e.g. json.Number → float64)
	// then through structpb for consistency with how terraform outputs are stored.
	b, err := json.Marshal(h.state.outputs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal pulumi outputs: %w", err)
	}

	var normalized map[string]interface{}
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, fmt.Errorf("unable to unmarshal pulumi outputs: %w", err)
	}

	spb, err := structpb.NewStruct(normalized)
	if err != nil {
		return nil, fmt.Errorf("unable to convert pulumi outputs to struct: %w", err)
	}

	return spb.AsMap(), nil
}
