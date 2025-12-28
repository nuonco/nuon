package examples

import (
	"go.temporal.io/sdk/workflow"
)

// QueryHandler demonstrates a query handler
// @temporal-gen-v2 query
func QueryHandler(ctx workflow.Context, input string) (string, error) {
	return "query result", nil
}

// GetStatusQuery demonstrates a query handler for getting status
// @temporal-gen-v2 query
func GetStatusQuery(ctx workflow.Context, input string) (string, error) {
	return "running", nil
}

// UpdateHandler demonstrates an update handler
// @temporal-gen-v2 update
// @id my-update-id
func UpdateHandler(ctx workflow.Context, input string) (string, error) {
	return "update result", nil
}
