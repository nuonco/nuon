package apps

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMinCLIMessage(t *testing.T) {
	current, err := semver.NewVersion("0.19.1000")
	require.NoError(t, err)
	minimum, err := semver.NewVersion("0.19.1102")
	require.NoError(t, err)

	assert.Equal(t,
		"Your CLI (0.19.1000) is older than the minimum 0.19.1102 this control plane syncs app configs for. Actions and runbooks will not be synced to installs. see https://docs.nuon.co/cli to update.",
		syncMinCLIMessage(current, minimum),
	)
}
