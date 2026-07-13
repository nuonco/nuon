package shell

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_NormalizesNilEnvVars(t *testing.T) {
	s, err := New(validator.New(), WithEnvVars(nil))
	require.NoError(t, err)
	assert.NotNil(t, s.EnvVars, "nil env vars must normalize to an empty map so setCredentials can merge into it")
}
