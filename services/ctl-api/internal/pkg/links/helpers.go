package links

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func configFromContext(ctx context.Context) *internal.Config {
	val := ctx.Value(keys.CfgCtxKey)
	valObj, ok := val.(*internal.Config)
	if !ok {
		return nil
	}

	return valObj
}

func isEmployeeFromContext(ctx context.Context) bool {
	isEmployee := ctx.Value(keys.IsEmployeeCtxKey)
	if isEmployee == nil {
		return false
	}

	return isEmployee.(bool)
}

// orgIDFromContext returns the org id from the context. Notably, this depends on the `middlewares/org` to set
// this, but we do not use that code to prevent a cycle import
func orgIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.OrgIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}
