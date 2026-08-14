package service

import (
	"strconv"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func provisionWorkflowMetadata(stackOnly bool) map[string]string {
	metadata := map[string]string{}
	if stackOnly {
		metadata[app.WorkflowMetadataKeyStackOnly] = strconv.FormatBool(true)
	}
	return metadata
}

func provisionPhaseDescription(stackOnly bool) string {
	if stackOnly {
		return "Setting up runner, skipping sandbox and components"
	}
	return "Setting up runner and sandbox resources"
}
