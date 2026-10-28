package procisolation

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRemovesTokenFromChildEnv(t *testing.T) {
	t.Setenv("RUNNER_API_TOKEN", "test-token")
	t.Setenv("RUNNER_ID", "runner-id")

	require.NoError(t, Apply())

	_, ok := os.LookupEnv("RUNNER_API_TOKEN")
	assert.False(t, ok)
	assert.Equal(t, "runner-id", os.Getenv("RUNNER_ID"))

	out, err := exec.Command("env").Output()
	require.NoError(t, err)
	assert.NotContains(t, string(out), "test-token")
}
