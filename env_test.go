package terraform

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_getEnv(t *testing.T) {
	env := getEnv()
	assert.NotNil(t, env)

	for _, envVar := range os.Environ() {
		pair := strings.SplitN(envVar, "=", 2)
		assert.Equal(t, pair[1], env[pair[0]])
	}
}
