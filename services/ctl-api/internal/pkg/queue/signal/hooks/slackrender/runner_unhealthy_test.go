package slackrender

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRunnerUnhealthyMessage(t *testing.T) {
	msg := BuildRunnerUnhealthyMessage(Event{
		Kind:       KindRunnerUnhealthy,
		Transition: TransitionUnhealthy,
		OrgID:      "org_1",
		OrgName:    "Acme",
		Workflow: WorkflowRef{
			OwnerID:   "ins_1",
			OwnerType: OwnerTypeInstalls,
			OwnerName: "Production",
		},
		Links: &ContextLinks{Install: "https://app.nuon.co/org_1/installs/ins_1"},
		Metadata: map[string]any{
			"runner_id":         "run_1",
			"runner_name":       "Default runner",
			"runner_group_type": "install",
			"from_status":       "active",
			"to_status":         "offline",
			"reason":            "no active install process",
		},
	})

	assert.Equal(t, "🚨 Runner unhealthy — Default runner", msg.Text)
	encoded, err := json.Marshal(msg.Blocks)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "Runner unhealthy · Default runner")
	assert.Contains(t, string(encoded), "active → offline")
	assert.Contains(t, string(encoded), "no active install process")
	assert.Contains(t, string(encoded), "Production")
}
