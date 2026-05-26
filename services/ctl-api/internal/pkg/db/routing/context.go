package routing

import "context"

type Pool string

const (
	PoolPrimary Pool = "primary"
	PoolReplica Pool = "replica"
)

type (
	replicaOptInKey  struct{}
	replicaOptOutKey struct{}
	replicaForceKey  struct{}
	decisionKey      struct{}
)

func WithReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, replicaOptInKey{}, true)
}

func UseReplica(ctx context.Context) bool {
	v, _ := ctx.Value(replicaOptInKey{}).(bool)
	return v
}

func WithoutReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, replicaOptOutKey{}, true)
}

func IsWithoutReplica(ctx context.Context) bool {
	v, _ := ctx.Value(replicaOptOutKey{}).(bool)
	return v
}

func WithForceReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, replicaForceKey{}, true)
}

func IsForceReplica(ctx context.Context) bool {
	v, _ := ctx.Value(replicaForceKey{}).(bool)
	return v
}

func WithDecision(ctx context.Context, p Pool) context.Context {
	return context.WithValue(ctx, decisionKey{}, p)
}

func DecisionFromContext(ctx context.Context) Pool {
	if v, ok := ctx.Value(decisionKey{}).(Pool); ok {
		return v
	}
	return PoolPrimary
}
