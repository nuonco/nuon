package runnerctx

import "context"

type runnerID struct{}

func FromContext(ctx context.Context) string {
	val := ctx.Value(runnerID{})
	if val == nil {
		return ""
	}

	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}
