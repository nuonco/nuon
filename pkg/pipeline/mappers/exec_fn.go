package mappers

import (
	"context"
	"fmt"
)

// GetExecFn returns the function that should be executed for a pipeline step
func (d *defaultMapper) GetExecFn(ctx context.Context, fn interface{}) (ExecFn, error) {
	switch f := fn.(type) {
	case execInitFn:
		return f.exec, nil
	case execPlanFn:
		return f.exec, nil
	case execOutputFn:
		return f.exec, nil
	case execStateFn:
		return f.exec, nil
	case ExecFn:
		return f, nil
	default:
	}

	return nil, fmt.Errorf("unable to map interface to an exec function: %T", fn)
}
