package slackrender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAwaitingRetryPresentation(t *testing.T) {
	assert.Equal(t, "Awaiting manual retry", transitionPhrase(TransitionAwaitingRetry))
	assert.Equal(t, "⚠️", statusEmoji(TransitionAwaitingRetry))
}
