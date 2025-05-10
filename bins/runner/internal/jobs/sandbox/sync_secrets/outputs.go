package terraform

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return generics.ToIntMap(h.state.outputs), nil
}
