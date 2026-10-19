package ocicopy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

func TestChildVersionTag(t *testing.T) {
	t.Parallel()

	platform := ocispec.Descriptor{Platform: &ocispec.Platform{OS: "linux", Architecture: "amd64"}}
	arm := ocispec.Descriptor{Platform: &ocispec.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}}
	attestation := ocispec.Descriptor{Platform: &ocispec.Platform{OS: "unknown", Architecture: "unknown"}}
	annotated := ocispec.Descriptor{Annotations: map[string]string{"vnd.docker.reference.type": attestationReferenceType}}

	seen := map[string]int{}
	first, ok := childVersionTag("v1.2.3", platform, seen)
	require.True(t, ok)
	require.Equal(t, "v1.2.3--linux-amd64", first)

	second, ok := childVersionTag("v1.2.3", platform, seen)
	require.True(t, ok)
	require.Equal(t, "v1.2.3--linux-amd64-2", second)

	variant, ok := childVersionTag("v1.2.3", arm, seen)
	require.True(t, ok)
	require.Equal(t, "v1.2.3--linux-arm64-v8", variant)

	att, ok := childVersionTag("v1.2.3", attestation, seen)
	require.True(t, ok)
	require.Equal(t, "v1.2.3--attestation", att)

	ann, ok := childVersionTag("v1.2.3", annotated, seen)
	require.True(t, ok)
	require.Equal(t, "v1.2.3--attestation-2", ann)

	_, ok = childVersionTag(strings.Repeat("a", maxTagLength), platform, map[string]int{})
	require.False(t, ok)
}

func TestTagIndexChildrenLeavesConflictingTag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dst := memory.New()

	platform := pushManifest(t, ctx, dst)
	platform.Platform = &ocispec.Platform{OS: "linux", Architecture: "amd64"}
	attestation := pushManifest(t, ctx, dst)
	attestation.Platform = &ocispec.Platform{OS: "unknown", Architecture: "unknown"}
	other := pushManifest(t, ctx, dst)
	require.NoError(t, dst.Tag(ctx, other, "v1.2.3--attestation"))

	root := pushIndex(t, ctx, dst, platform, attestation)
	require.NoError(t, tagIndexChildren(ctx, dst, root, "v1.2.3"))

	gotPlatform, err := dst.Resolve(ctx, "v1.2.3--linux-amd64")
	require.NoError(t, err)
	require.Equal(t, platform.Digest, gotPlatform.Digest)

	gotAttestation, err := dst.Resolve(ctx, "v1.2.3--attestation")
	require.NoError(t, err)
	require.Equal(t, other.Digest, gotAttestation.Digest)

	require.NoError(t, tagIndexChildren(ctx, dst, root, "v1.2.3"))
	gotPlatform, err = dst.Resolve(ctx, "v1.2.3--linux-amd64")
	require.NoError(t, err)
	require.Equal(t, platform.Digest, gotPlatform.Digest)
}

func TestTagIndexChildrenSkipsDigestReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dst := memory.New()

	platform := pushManifest(t, ctx, dst)
	platform.Platform = &ocispec.Platform{OS: "linux", Architecture: "amd64"}
	root := pushIndex(t, ctx, dst, platform)

	require.NoError(t, tagIndexChildren(ctx, dst, root, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	_, err := dst.Resolve(ctx, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa--linux-amd64")
	require.Error(t, err)
}

func pushManifest(t *testing.T, ctx context.Context, store *memory.Store) ocispec.Descriptor {
	t.Helper()
	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, "application/vnd.oci.image.config.v1+json", oras.PackManifestOptions{})
	require.NoError(t, err)
	return desc
}

func pushIndex(t *testing.T, ctx context.Context, store *memory.Store, manifests ...ocispec.Descriptor) ocispec.Descriptor {
	t.Helper()
	payload, err := json.Marshal(ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: manifests,
	})
	require.NoError(t, err)
	root, err := oras.PushBytes(ctx, store, ocispec.MediaTypeImageIndex, payload)
	require.NoError(t, err)
	return root
}
