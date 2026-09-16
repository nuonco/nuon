package syncer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBranchSyncOption(t *testing.T) {
	t.Run("enabled by default", func(t *testing.T) {
		s := NewDBSyncer(nil, nil, nil, nil, nil, nil, nil, nil, "", nil, "").(*syncer)

		require.True(t, s.syncBranches)
	})

	t.Run("can be disabled", func(t *testing.T) {
		s := NewDBSyncer(nil, nil, nil, nil, nil, nil, nil, nil, "", nil, "", WithoutBranchSync()).(*syncer)

		require.False(t, s.syncBranches)
	})
}
