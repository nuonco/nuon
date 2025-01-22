package terraform

import (
	"context"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/outputs"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return outputs.TerraformOutputs(ctx, h.state.tfWorkspace)
}
