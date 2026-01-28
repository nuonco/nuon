package imagemetadata

import (
	"context"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	if h.state.metadata == nil {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		"digest":             h.state.metadata.Digest,
		"signed":             h.state.metadata.Signed,
		"has_sbom":           h.state.metadata.SBOM != nil && h.state.metadata.SBOM.Present,
		"attestations_count": len(h.state.metadata.Attestations),
	}, nil
}
