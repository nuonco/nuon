package generics

import (
	"testing"

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
	assert.Equal(t, len(obj.ValidShortID), 26)
}
