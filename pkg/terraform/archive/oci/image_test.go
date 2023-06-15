package oci

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestImage_RepoURL(t *testing.T) {
	img := generics.GetFakeObj[Image]()

	repoURL := img.RepoURL()
	assert.Contains(t, repoURL, img.Repo)
	assert.Contains(t, repoURL, img.Registry)
}
