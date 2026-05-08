package routing

import "context"

type contextKey struct{}

// WithReplica returns a new context that signals read queries should be routed
// to the replica database connection.
func WithReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, true)
}

// UseReplica reports whether the context has been marked for replica routing.
func UseReplica(ctx context.Context) bool {
	v, _ := ctx.Value(contextKey{}).(bool)
	return v
}
