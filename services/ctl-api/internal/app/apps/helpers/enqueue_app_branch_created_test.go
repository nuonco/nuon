package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnqueueAppBranchCreatedIfFirstRequiresIDs(t *testing.T) {
	h := &Helpers{}
	require.Error(t, h.EnqueueAppBranchCreatedIfFirst(context.Background(), "", "cfg-1"))
	require.Error(t, h.EnqueueAppBranchCreatedIfFirst(context.Background(), "branch-1", ""))
}
