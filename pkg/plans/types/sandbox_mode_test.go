package plantypes

import (
	_ "crypto/sha256"
	"testing"

	"github.com/distribution/reference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The sandbox ref is parsed as a container reference, not just interpolated
// into a template, so a decorative placeholder digest silently fails every
// image-backed action in sandbox mode.
func TestFakeOCISyncOutputsRefIsDigestPinned(t *testing.T) {
	out := FakeOCISyncOutputs("registry.example.com/nuon/app-service", "v1.2.3")

	image, ok := out["image"].(map[string]any)
	require.True(t, ok, "outputs must carry an image map")

	ref, ok := image["ref"].(string)
	require.True(t, ok, "image outputs must carry a ref")

	named, err := reference.ParseDockerRef(ref)
	require.NoError(t, err, "ref %q must parse as a container reference", ref)

	_, digested := named.(reference.Digested)
	assert.True(t, digested, "ref %q must be digest-pinned", ref)

	assert.Equal(t, "registry.example.com/nuon/app-service", image["repository"])
	assert.Equal(t, "v1.2.3", image["tag"])
}

// A public source image is only ever present at its tag, since the sandbox
// skips the copy that would give the placeholder digest any meaning.
func TestFakePublicOCISyncOutputsRefIsTagged(t *testing.T) {
	out := FakePublicOCISyncOutputs("us-west1-docker.pkg.dev/acme/tools/dns-check", "1.0.0")

	image, ok := out["image"].(map[string]any)
	require.True(t, ok, "outputs must carry an image map")

	ref, ok := image["ref"].(string)
	require.True(t, ok, "image outputs must carry a ref")

	named, err := reference.ParseDockerRef(ref)
	require.NoError(t, err, "ref %q must parse as a container reference", ref)

	_, digested := named.(reference.Digested)
	assert.False(t, digested, "ref %q must not be pinned to a digest the source lacks", ref)

	tagged, ok := named.(reference.Tagged)
	require.True(t, ok, "ref %q must carry a tag", ref)
	assert.Equal(t, "1.0.0", tagged.Tag())

	assert.Equal(t, "us-west1-docker.pkg.dev/acme/tools/dns-check", image["repository"])
}
