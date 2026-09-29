package containerimage

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/generics"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	// why: `repository` and `tag` are a long-standing public output contract:
	// app configs compose them as `{{.repository}}:{{.tag}}` (helm values,
	// terraform vars). They must stay the bare repository and resolved tag —
	// rewriting repository to the digest-pinned form breaks every such
	// template, and clearing the tag does NOT fail loudly: charts default an
	// empty tag to .Chart.AppVersion, yielding invalid `repo@sha256:...:tag`
	// refs that only surface as pull failures at deploy time.
	//
	// Digest pinning is exposed additively instead:
	//
	//   - ref         → "<full_repo>@sha256:<digest>" for templates that
	//                   opt in to digest-pinned pulls ("" without a digest)
	//   - display_tag → the resolved tag, stable alias of `tag`
	digest := ""
	if h.state.descriptor != nil {
		digest = h.state.descriptor.Digest.String()
	}

	fullRepo := strings.TrimPrefix(strings.TrimPrefix(h.state.plan.Dst.Repository, "https://"), "http://")
	loginServer := strings.TrimPrefix(strings.TrimPrefix(h.state.plan.Dst.LoginServer, "https://"), "http://")
	if loginServer != "" && fullRepo != "" && !strings.HasPrefix(fullRepo, loginServer+"/") {
		fullRepo = fmt.Sprintf("%s/%s", loginServer, fullRepo)
	}

	refOut := ""
	if digest != "" && fullRepo != "" {
		refOut = fmt.Sprintf("%s@%s", fullRepo, digest)
	}

	obj := map[string]interface{}{
		"repository":  fullRepo,
		"tag":         h.state.plan.DstTag,
		"ref":         refOut,
		"display_tag": h.state.plan.DstTag,
	}

	if h.state.descriptor != nil {
		obj = generics.MergeMap(obj, map[string]any{
			"media_type":    h.state.descriptor.MediaType,
			"digest":        digest,
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
