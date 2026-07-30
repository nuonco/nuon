package worker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanupTriggerEventsNeedsContinuation(t *testing.T) {
	require.False(t, cleanupTriggerEventsNeedsContinuation(cleanupTriggerEventMaxBatches-1, cleanupTriggerEventBatchSize))
	require.False(t, cleanupTriggerEventsNeedsContinuation(cleanupTriggerEventMaxBatches, cleanupTriggerEventBatchSize-1))
	require.True(t, cleanupTriggerEventsNeedsContinuation(cleanupTriggerEventMaxBatches, cleanupTriggerEventBatchSize))
}
