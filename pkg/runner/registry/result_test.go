package registry

import (
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestToAPIResultIncludesDestinationDescriptor(t *testing.T) {
	desc := &ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromString("bundle artifact"),
		Size:      1234,
	}

	result := ToAPIResult("registry.example.com/org/app", desc)

	require.True(t, result.Success)
	require.Equal(t, "registry.example.com/org/app", result.OutputRepository)
	require.Equal(t, desc.Digest.String(), result.OutputDigest)
	require.Equal(t, desc.MediaType, result.OutputMediaType)
	require.Equal(t, desc.Size, result.OutputSize)
}
