package app

import (
	"context"
)

// createdByIDFromContext returns the user id from the context. Notably, this depends on the `middlewares/auth` to set
// this, but we do not use that code to prevent a cycle import
func createdByIDFromContext(ctx context.Context) string {
	val := ctx.Value("account_id")
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

// orgIDFromContext returns the org id from the context. Notably, this depends on the `middlewares/org` to set
// this, but we do not use that code to prevent a cycle import
func orgIDFromContext(ctx context.Context) string {
	val := ctx.Value("org_id")
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}
