package ocicopy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

func TestCopyWithReferrers(t *testing.T) {
	ctx := context.Background()
	src := memory.New()
	dst := memory.New()

	platformImage, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/vnd.oci.image.config.v1+json", oras.PackManifestOptions{})
	require.NoError(t, err)
	embeddedAttestation, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/vnd.in-toto+json", oras.PackManifestOptions{})
	require.NoError(t, err)
	embeddedAttestation.Platform = &ocispec.Platform{OS: "unknown", Architecture: "unknown"}

	indexJSON, err := json.Marshal(ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: []ocispec.Descriptor{platformImage, embeddedAttestation},
	})
	require.NoError(t, err)
	image, err := oras.PushBytes(ctx, src, ocispec.MediaTypeImageIndex, indexJSON)
	require.NoError(t, err)
	require.NoError(t, src.Tag(ctx, image, "latest"))

	sbom, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/spdx+json", oras.PackManifestOptions{
		Subject: &image,
	})
	require.NoError(t, err)

	signature, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/vnd.dev.cosign.simplesigning.v1+json", oras.PackManifestOptions{
		Subject: &image,
	})
	require.NoError(t, err)

	attestation, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/vnd.in-toto+json", oras.PackManifestOptions{
		Subject: &sbom,
	})
	require.NoError(t, err)

	legacyArtifacts := make(map[string]ocispec.Descriptor)
	legacyTagPrefix := "sha256-" + image.Digest.Encoded()
	for _, artifact := range []struct {
		suffix       string
		artifactType string
	}{
		{suffix: ".sig", artifactType: "application/vnd.dev.cosign.simplesigning.v1+json"},
		{suffix: ".att", artifactType: "application/vnd.dsse.envelope.v1+json"},
		{suffix: ".sbom", artifactType: "application/spdx+json"},
	} {
		desc, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, artifact.artifactType, oras.PackManifestOptions{})
		require.NoError(t, err)
		tag := legacyTagPrefix + artifact.suffix
		require.NoError(t, src.Tag(ctx, desc, tag))
		legacyArtifacts[tag] = desc
	}

	got, err := copyWithReferrers(ctx, src, "latest", dst, "copied", oras.DefaultCopyGraphOptions)
	require.NoError(t, err)
	require.Equal(t, image, got)

	for _, artifact := range []struct {
		name string
		desc ocispec.Descriptor
	}{
		{name: "embedded attestation manifest", desc: embeddedAttestation},
		{name: "SBOM", desc: sbom},
		{name: "signature", desc: signature},
		{name: "nested attestation", desc: attestation},
	} {
		t.Run(artifact.name, func(t *testing.T) {
			exists, err := dst.Exists(ctx, artifact.desc)
			require.NoError(t, err)
			require.True(t, exists)
		})
	}

	for tag, want := range legacyArtifacts {
		t.Run("legacy "+tag, func(t *testing.T) {
			got, err := dst.Resolve(ctx, tag)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
