package interests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestRunnerUnhealthyInterests(t *testing.T) {
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeRunnerUnhealthy,
		Phase:      signal.SignalPhaseExecute,
	}
	outcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}

	require.NoError(t, Validate(Interests{Resources: map[ResourceKind]ResourceCfg{
		ResourceRunners: {Ops: []string{"unhealthy"}, Outcome: OutcomeFailures},
	}}))
	assert.Equal(t, []string{
		"resource:runners",
		"op:runners.unhealthy",
		SlugEventRunnerUnhealthy,
		SlugOutcomeCompletion,
		SlugOutcomeFailures,
	}, Classify(event, outcome, nil))

	assert.True(t, Matches(event, outcome, nil, Interests{Resources: map[ResourceKind]ResourceCfg{
		ResourceRunners: {Ops: []string{"unhealthy"}, Outcome: OutcomeFailures},
	}}))
	assert.False(t, Matches(event, outcome, nil, Interests{Resources: map[ResourceKind]ResourceCfg{
		ResourceRunners: {Ops: []string{"inactive"}, Outcome: OutcomeAll},
	}}))
	assert.False(t, Matches(event, outcome, nil, Interests{Resources: map[ResourceKind]ResourceCfg{
		ResourceRunners: {Ops: []string{"unhealthy"}, Outcome: OutcomeNone},
	}}))
}
