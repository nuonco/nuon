package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPhoneHomeScriptRaw(t *testing.T) {
	a := Activities{}
	script, err := a.GetPhoneHomeScriptRaw(context.TODO(), &GetPhoneHomeScriptRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, script, "script should not be empty")
}
