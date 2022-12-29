package orgcontext

import (
	"context"
	"fmt"
)

var (
	errNotFound = fmt.Errorf("org context was not found")
	errInvalid  = fmt.Errorf("org context was invalid")
)

func Get(ctx context.Context) (*Context, error) {
	val := ctx.Value(orgContextKey{})
	if val == nil {
		return nil, errNotFound
	}

	orgCtx, ok := val.(*Context)
	if !ok {
		return nil, errInvalid
	}

	return orgCtx, nil
}
