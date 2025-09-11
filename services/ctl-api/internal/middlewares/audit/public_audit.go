package audit

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"

type publicMiddleware struct {
	*baseMiddleware
}

func (m *publicMiddleware) Name() string {
	return "public_audit"
}

func NewPublic(params Params) *publicMiddleware {
	return &publicMiddleware{
		baseMiddleware: newBaseMiddleware(params, api.APIContextTypePublic.String()),
	}
}
