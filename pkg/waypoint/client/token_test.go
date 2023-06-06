package client

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTokenSecretName(t *testing.T) {
	id := shortid.NewNanoID("") //prefix=def
	token := DefaultTokenSecretName(id)
	assert.Contains(t, token, id)
}
