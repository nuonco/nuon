package containerimage

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/generics"
)

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	// When the runner has resolved a manifest digest, the `repository`,
	// `tag`, and `ref` outputs are rewritten so user templates render
	// digest-pinned references by default:
	//
	//   - repository  → "<full_repo>@sha256:<digest>" (existing
	//                   `image: {{.repository}}` templates pin automatically)
	//   - tag         → "" (a legacy
	//                   `image: {{.repository}}:{{.tag}}` template surfaces
	//                   as an invalid trailing-colon ref instead of
	//                   silently deploying a mutable tag)
	//   - ref         → same digest-pinned form, exposed as a stable alias
	//   - display_tag → the resolved tag for human display
	//
	// Without a digest (non-image artifacts) the legacy `repository = bare
	// repo`, `tag = DstTag` shape is preserved.
	digest := ""
	if h.state.descriptor != nil {
		digest = h.state.descriptor.Digest.String()
	}

	// Mirror the defensive prepend logic in
	// pkg/registry/access_info.go::RepositoryURI: only prefix LoginServer
	// when Repository doesn't already include it. Required for registries
	// like GAR (and ECR) where the configured Repository value is the
	// fully-qualified `<host>/<path>` form — an unconditional prepend
	// produces broken `<host>/<host>/<path>` refs that registries reject.
	fullRepo := strings.TrimPrefix(strings.TrimPrefix(h.state.plan.Dst.Repository, "https://"), "http://")
	loginServer := strings.TrimPrefix(strings.TrimPrefix(h.state.plan.Dst.LoginServer, "https://"), "http://")
	if loginServer != "" && fullRepo != "" && !strings.HasPrefix(fullRepo, loginServer+"/") {
		fullRepo = fmt.Sprintf("%s/%s", loginServer, fullRepo)
	}

	var (
		repositoryOut string
		tagOut        string
		refOut        string
	)
	if digest != "" && fullRepo != "" {
		refOut = fmt.Sprintf("%s@%s", fullRepo, digest)
		repositoryOut = refOut
		tagOut = ""
	} else {
		repositoryOut = h.state.plan.Dst.Repository
		tagOut = h.state.plan.DstTag
	}

	obj := map[string]interface{}{
		"repository":  repositoryOut,
		"tag":         tagOut,
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
