package executeflow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCallbacksWorkflowID(t *testing.T) {
	sig := &Signal{WorkflowID: "wfl_1"}
	assert.Empty(t, sig.CompletionCallbacksWorkflowID())

	sig.Resident = true
	assert.Equal(t, "wfl_1", sig.CompletionCallbacksWorkflowID())
}

func TestNewSignalIsResident(t *testing.T) {
	sig := NewSignal("wfl_1")
	require.Equal(t, "wfl_1", sig.WorkflowID)
	require.True(t, sig.Resident)
}

func TestLegacySignalDoesNotGateCompletionCallbacks(t *testing.T) {
	var sig Signal
	require.NoError(t, json.Unmarshal([]byte(`{"workflow_id":"wfl_1"}`), &sig))
	assert.Empty(t, sig.CompletionCallbacksWorkflowID())
}
