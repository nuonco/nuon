package shortid

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

type testFakeObj struct {
	ShortID string `faker:"shortID"`
}

func TestGetFakeObj(t *testing.T) {
	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)
	assert.Equal(t, len(obj.ShortID), 26)
}
