package oci

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/generics"
)

func TestImage_RepoURL(t *testing.T) {
	img := generics.GetFakeObj[Image]()

	repoURL := img.RepoURL()
	assert.Contains(t, repoURL, img.Repo)
	assert.Contains(t, repoURL, img.Registry)
}
