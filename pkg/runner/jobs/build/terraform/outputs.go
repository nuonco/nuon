package terraform

import (
	"context"

	"github.com/pkg/errors"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	srcFiles, err := h.getSourceFiles(ctx, h.state.workspace.Source().AbsPath())
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source files")
	}

	return map[string]interface{}{
		"files": srcFiles,
		"image": map[string]interface{}{
			"tag": h.state.resultTag,
		},
	}, nil
}
