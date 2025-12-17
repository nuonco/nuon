package audit

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"

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
