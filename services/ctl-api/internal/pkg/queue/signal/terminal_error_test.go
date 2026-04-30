package signal

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/temporal"
)

func TestNewTerminalError(t *testing.T) {
	err := NewTerminalError(
		"no_component_build",
		"Ensure there is an active build for this component (id %s) before retrying.",
		"abc123",
	)
	assert.True(t, IsTerminalError(err))
	assert.Equal(t, "no_component_build", TerminalReasonCode(err))
	assert.Contains(t, err.Error(), "Ensure there is an active build")
	assert.Contains(t, err.Error(), "abc123")
}

func TestWrapTerminal(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, WrapTerminal(nil, "any"))
	})

	t.Run("preserves underlying message", func(t *testing.T) {
		base := errors.New("disk full")
		wrapped := WrapTerminal(base, "infra_unavailable")
		assert.True(t, IsTerminalError(wrapped))
		assert.Equal(t, "infra_unavailable", TerminalReasonCode(wrapped))
		assert.Contains(t, wrapped.Error(), "disk full")
	})
}

func TestIsTerminalError_NilAndPlain(t *testing.T) {
	assert.False(t, IsTerminalError(nil))
	assert.False(t, IsTerminalError(errors.New("just a regular error")))
	assert.Equal(t, "", TerminalReasonCode(nil))
	assert.Equal(t, "", TerminalReasonCode(errors.New("just a regular error")))
}

// TestIsTerminalError_SurvivesQueueHandlerWrapping verifies that the marker
// remains detectable after the queue handler / AwaitSignal / step executor
// flatten the error into a fresh non-retryable ApplicationError carrying the
// human description, then wrap it again with errors.Wrap.
//
// This mirrors the chain in:
//   - handler/execute.go (humanDesc → new ApplicationError)
//   - client/await_signal.go (NewNonRetryableApplicationError(description, "SIGNAL_FAILED", nil))
//   - executeworkflowstep/execute_step.go (errors.Wrapf(...))
func TestIsTerminalError_SurvivesQueueHandlerWrapping(t *testing.T) {
	innerErr := NewTerminalError(
		"no_component_build",
		"Ensure there is an active build for this component before retrying.",
	)

	humanDesc := HumanError(innerErr)
	rebuilt := temporal.NewNonRetryableApplicationError(humanDesc, "SIGNAL_FAILED", nil)
	stepErr := errors.Wrap(rebuilt, "queue signal execution failed for step deploy")

	assert.True(t, IsTerminalError(stepErr), "marker must survive the full wrapping chain")
	assert.Equal(t, "no_component_build", TerminalReasonCode(stepErr))
}
