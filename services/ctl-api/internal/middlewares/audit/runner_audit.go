package audit

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"

type runnerMiddleware struct {
	*baseMiddleware
}

func (m *runnerMiddleware) Name() string {
	return "runner_audit"
}

func NewRunner(params Params) *runnerMiddleware {
	return &runnerMiddleware{
		baseMiddleware: newBaseMiddleware(params, api.APIContextTypeRunner.String()),
	}
}
