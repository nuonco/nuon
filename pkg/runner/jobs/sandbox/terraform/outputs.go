package terraform

import (
	"context"

	"github.com/nuonco/nuon/pkg/runner/outputs"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return outputs.TerraformOutputs(ctx, h.state.tfWorkspace)
}
