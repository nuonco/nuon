package generics

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/generics/fakers"
	"github.com/stretchr/testify/assert"
)

type testFakeObj struct {
	ValidShortID   string `faker:"shortID"`
	InvalidShortID string `faker:"len=26"`
}

func TestGetFakeObj(t *testing.T) {
	fakers.Register()
	obj := GetFakeObj[testFakeObj]()

	parsed, err := shortid.ToUUID(obj.ValidShortID)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed)

	parsed, err = shortid.ToUUID(obj.InvalidShortID)
	assert.Error(t, err)
	assert.Empty(t, parsed)
}
