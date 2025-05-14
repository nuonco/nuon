package containerimage

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	obj := map[string]interface{}{
		"tag":        h.state.plan.DstTag,
		"repository": h.state.plan.Dst.Repository,
	}

	if h.state.descriptor != nil {
		obj = generics.MergeMap(obj, map[string]any{
			"media_type":    h.state.descriptor.MediaType,
			"digest":        h.state.descriptor.Digest.String(),
			"size":          h.state.descriptor.Size,
			"urls":          h.state.descriptor.URLs,
			"annotations":   h.state.descriptor.Annotations,
			"artifact_type": h.state.descriptor.ArtifactType,
		})
	}
	if h.state.descriptor != nil && h.state.descriptor.Platform != nil {
		obj = generics.MergeMap(obj, map[string]any{
			"platform": map[string]any{
				"architecture": h.state.descriptor.Platform.Architecture,
				"os":           h.state.descriptor.Platform.OS,
				"os_version":   h.state.descriptor.Platform.OSVersion,
				"variant":      h.state.descriptor.Platform.Variant,
				"os_features":  h.state.descriptor.Platform.OSFeatures,
			},
		})
	}

	return map[string]any{
		"image": obj,
	}, nil
}
