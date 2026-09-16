package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestStepScriptContents(t *testing.T) {
	t.Run("wraps command as a posix script", func(t *testing.T) {
		got, err := stepScriptContents(&models.AppActionWorkflowStepConfig{
			Command: "migrate -database \"$DATABASE_URL\" up",
		})
		require.NoError(t, err)
		assert.Equal(t, "#!/bin/sh\nmigrate -database \"$DATABASE_URL\" up\n", got)
	})

	t.Run("preserves inline_contents shebang", func(t *testing.T) {
		got, err := stepScriptContents(&models.AppActionWorkflowStepConfig{
			InlineContents: "#!/usr/bin/env sh\necho hi\n",
		})
		require.NoError(t, err)
		assert.Equal(t, "#!/usr/bin/env sh\necho hi\n", got)
	})

	t.Run("adds a shebang to inline_contents without one", func(t *testing.T) {
		got, err := stepScriptContents(&models.AppActionWorkflowStepConfig{
			InlineContents: "echo hi",
		})
		require.NoError(t, err)
		assert.Equal(t, "#!/bin/sh\necho hi", got)
	})

	t.Run("prefers inline_contents when both are set", func(t *testing.T) {
		got, err := stepScriptContents(&models.AppActionWorkflowStepConfig{
			InlineContents: "echo inline",
			Command:        "echo command",
		})
		require.NoError(t, err)
		assert.Equal(t, "#!/bin/sh\necho inline", got)
	})

	t.Run("errors when neither command nor inline_contents is set", func(t *testing.T) {
		_, err := stepScriptContents(&models.AppActionWorkflowStepConfig{})
		require.Error(t, err)
	})
}

func TestIsUnderWorkspace(t *testing.T) {
	assert.True(t, isUnderWorkspace("/tmp/work", "/tmp/work/actions/healthcheck"))
	assert.False(t, isUnderWorkspace("/tmp/work", "sh"))
	assert.False(t, isUnderWorkspace("/tmp/work", "/tmp/other/healthcheck"))
	assert.False(t, isUnderWorkspace("/tmp/work", "/tmp/work-evil/healthcheck"))
}
