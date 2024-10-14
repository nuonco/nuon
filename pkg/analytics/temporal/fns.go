package temporalanalytics

import "go.temporal.io/sdk/workflow"

type (
	UserIDFn func(workflow.Context) (string, error)
)
