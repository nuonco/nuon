package links

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
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

func orgIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.OrgIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}
