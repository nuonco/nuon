package installs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestPriorDeployBuildID(t *testing.T) {
	t.Run("uses prior deploy build", func(t *testing.T) {
		id, ok := priorDeployBuildID(&models.AppInstallDeploy{BuildID: "bld123"}, nil)
		require.True(t, ok)
		assert.Equal(t, "bld123", id)
	})

	t.Run("missing prior deploy falls through", func(t *testing.T) {
		id, ok := priorDeployBuildID(nil, errors.New("record not found"))
		require.False(t, ok)
		assert.Empty(t, id)
	})

	t.Run("empty prior deploy falls through", func(t *testing.T) {
		id, ok := priorDeployBuildID(&models.AppInstallDeploy{}, nil)
		require.False(t, ok)
		assert.Empty(t, id)
	})
}
