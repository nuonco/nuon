package procisolation

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRemovesRunnerEnvFromChild(t *testing.T) {
	t.Setenv("RUNNER_API_TOKEN", "test-token")
	t.Setenv("RUNNER_ID", "runner-id")
	t.Setenv("RUNNER_API_URL", "https://runner.example.com")
	t.Setenv("PATH", os.Getenv("PATH"))

	require.NoError(t, Apply())

	for _, name := range []string{"RUNNER_API_TOKEN", "RUNNER_ID", "RUNNER_API_URL"} {
		_, ok := os.LookupEnv(name)
		assert.False(t, ok, name)
	}
	assert.NotEmpty(t, os.Getenv("PATH"))

	out, err := exec.Command("env").Output()
	require.NoError(t, err)
	assert.NotContains(t, string(out), "test-token")
	assert.NotContains(t, string(out), "RUNNER_ID=")
	assert.NotContains(t, string(out), "runner.example.com")
}
