package orgiam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGARRepositoryResource(t *testing.T) {
	resource, err := garRepositoryResource("us-west1-docker.pkg.dev/my-project/my-repo")
	require.NoError(t, err)
	assert.Equal(t, "projects/my-project/locations/us-west1/repositories/my-repo", resource)

	resource, err = garRepositoryResource("europe-west4-docker.pkg.dev/proj/repo/")
	require.NoError(t, err)
	assert.Equal(t, "projects/proj/locations/europe-west4/repositories/repo", resource)

	_, err = garRepositoryResource("us-west1-docker.pkg.dev/my-project")
	assert.Error(t, err)

	_, err = garRepositoryResource("gcr.io/my-project/my-repo")
	assert.Error(t, err)
}
