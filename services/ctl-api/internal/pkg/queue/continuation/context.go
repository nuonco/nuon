package continuation

import "go.temporal.io/sdk/workflow"

type scopeKey struct{}

func WithScope(ctx workflow.Context, id string) workflow.Context {
	return workflow.WithValue(ctx, scopeKey{}, id)
}

func Scope(ctx workflow.Context) string {
	id, _ := ctx.Value(scopeKey{}).(string)
	return id
}
