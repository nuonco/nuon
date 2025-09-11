package audit

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"

type internalMiddleware struct {
	*baseMiddleware
}

func (m *internalMiddleware) Name() string {
	return "internal_audit"
}

func NewInternal(params Params) *internalMiddleware {
	return &internalMiddleware{
		baseMiddleware: newBaseMiddleware(params, api.APIContextTypeInternal.String()),
	}
}
