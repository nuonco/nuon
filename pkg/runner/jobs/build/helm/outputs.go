package helm

import (
	"context"

	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"files": []ociarchive.FileRef{
			{
				RelPath: defaultChartPackageFilename,
			},
		},
		"image": map[string]interface{}{
			"tag": h.state.resultTag,
		},
	}, nil
}
