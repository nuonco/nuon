package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsActionExecutionRunnerJob(t *testing.T) {
	assert.False(t, isActionExecutionRunnerJob(nil))
	assert.False(t, isActionExecutionRunnerJob(&RunnerJob{Type: RunnerJobTypeOCISync}))
	assert.True(t, isActionExecutionRunnerJob(&RunnerJob{Type: RunnerJobTypeActionsWorkflowRun}))
}
