package terraform

import (
	"context"

	"github.com/pkg/errors"

	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	srcFiles, err := ociarchive.SourceFiles(ctx, h.state.archiveSource, defaultFileType)
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
