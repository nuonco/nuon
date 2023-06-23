package generics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testFakeObj struct {
	String26 string `faker:"len=26"`
}

func TestGetFakeObj(t *testing.T) {
	obj := GetFakeObj[testFakeObj]()
	assert.Equal(t, len(obj.String26), 26)
}
