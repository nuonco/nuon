package registry

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestImageMapper(t *testing.T) {
	img := generics.GetFakeObj[*Image]()
	dockerImg := ImageMapper(img)

	assert.Equal(t, dockerImg.Image, img.Image)
	assert.Equal(t, dockerImg.Tag, img.Tag)
}
