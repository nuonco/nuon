package authz

import "context"

func SetAuthorityContext(ctx context.Context, authority *Authority) context.Context {
	return context.WithValue(ctx, authorityKey, authority)
}
