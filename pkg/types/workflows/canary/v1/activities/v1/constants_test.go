package activitiesv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivities(t *testing.T) {
	assert.Equal(t, Activity_ACTIVITY_POLL_WORKFLOW.String(), "ACTIVITY_POLL_WORKFLOW")
	assert.Equal(t, Activity_ACTIVITY_START_WORKFLOW.String(), "ACTIVITY_START_WORKFLOW")
}
