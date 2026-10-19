package ocicopy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

const (
	dockerManifestListMediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
	attestationReferenceType    = "attestation-manifest"
	childTagSeparator           = "--"
	maxTagLength                = 128
)

// tagIndexChildren gives each manifest inside an image index its own tag derived
// from the destination tag. Registries store those children by digest, so they
// otherwise show up as untagged even though the tagged index still references
// them. An existing tag that names a different digest is left alone: moving it
// would replace another image on a mutable registry and fail on an immutable one.
func tagIndexChildren(ctx context.Context, dst oras.Target, root ocispec.Descriptor, dstRef string) error {
	if !isImageIndex(root.MediaType) || !validTag(dstRef) {
		return nil
	}

	data, err := content.FetchAll(ctx, dst, root)
	if err != nil {
		return errors.Wrap(err, "fetch image index")
	}
	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return errors.Wrap(err, "decode image index")
	}

	l := pkgctx.LoggerOrDefault(ctx, zap.NewNop())
	seen := map[string]int{}
	for _, child := range index.Manifests {
		if child.Digest == "" {
			continue
		}
		tag, ok := childVersionTag(dstRef, child, seen)
		if !ok {
			l.Info("skipping child manifest tag",
				zap.String("parent_tag", dstRef),
				zap.String("digest", child.Digest.String()),
			)
			continue
		}

		existing, err := dst.Resolve(ctx, tag)
		if err == nil {
			if existing.Digest == child.Digest {
				continue
			}
			l.Info("leaving existing tag in place",
				zap.String("tag", tag),
				zap.String("existing_digest", existing.Digest.String()),
				zap.String("child_digest", child.Digest.String()),
			)
			continue
		}
		if !errors.Is(err, errdef.ErrNotFound) {
			return errors.Wrapf(err, "resolve child tag %s", tag)
		}
		if err := dst.Tag(ctx, child, tag); err != nil {
			l.Warn("unable to tag child manifest",
				zap.String("tag", tag),
				zap.String("digest", child.Digest.String()),
				zap.Error(err),
			)
			continue
		}
		l.Info("tagged child manifest",
			zap.String("tag", tag),
			zap.String("digest", child.Digest.String()),
		)
	}
	return nil
}

func isImageIndex(mediaType string) bool {
	return mediaType == ocispec.MediaTypeImageIndex || mediaType == dockerManifestListMediaType
}

func childVersionTag(parent string, desc ocispec.Descriptor, seen map[string]int) (string, bool) {
	base := parent + childTagSeparator + childTagSuffix(desc)
	n := seen[base]
	seen[base] = n + 1
	tag := base
	if n > 0 {
		tag = fmt.Sprintf("%s-%d", base, n+1)
	}
	if !validTag(tag) {
		return "", false
	}
	return tag, true
}

func childTagSuffix(desc ocispec.Descriptor) string {
	if isAttestation(desc) {
		return "attestation"
	}
	if desc.Platform != nil && desc.Platform.OS != "" && desc.Platform.Architecture != "" {
		suffix := sanitizeTagToken(desc.Platform.OS + "-" + desc.Platform.Architecture)
		if desc.Platform.Variant != "" {
			suffix += "-" + sanitizeTagToken(desc.Platform.Variant)
		}
		if suffix != "" {
			return suffix
		}
	}
	return "manifest"
}

func isAttestation(desc ocispec.Descriptor) bool {
	if desc.Annotations["vnd.docker.reference.type"] == attestationReferenceType {
		return true
	}
	return desc.Platform != nil && desc.Platform.OS == "unknown" && desc.Platform.Architecture == "unknown"
}

func sanitizeTagToken(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if isTagByte(value[i], false) {
			b.WriteByte(value[i])
			continue
		}
		b.WriteByte('-')
	}
	return strings.Trim(b.String(), "-.")
}

func validTag(tag string) bool {
	if tag == "" || len(tag) > maxTagLength {
		return false
	}
	for i := 0; i < len(tag); i++ {
		if !isTagByte(tag[i], i == 0) {
			return false
		}
	}
	return true
}

func isTagByte(b byte, first bool) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9', b == '_':
		return true
	case !first && (b == '.' || b == '-'):
		return true
	default:
		return false
	}
}
