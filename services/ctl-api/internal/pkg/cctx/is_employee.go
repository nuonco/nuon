package cctx

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func IsEmployeeFromContext(ctx ValueContext) (bool, error) {
	isEmployee := ctx.Value(keys.IsEmployeeCtxKey)
	if isEmployee == nil {
		return false, fmt.Errorf("is_employee is not set on the context")
	}

	return isEmployee.(bool), nil
}

func SetIsEmployeeWorkflowContext(ctx workflow.Context, isEmployee bool) workflow.Context {
	return workflow.WithValue(ctx, keys.IsEmployeeCtxKey, isEmployee)
}

func SetIsEmployeeContext(ctx context.Context, isEmployee bool) context.Context {
	return context.WithValue(ctx, keys.IsEmployeeCtxKey, isEmployee)
}
