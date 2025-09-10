package audit

type runnerMiddleware struct {
	baseMiddleware
}

func (m *runnerMiddleware) Name() string {
	return "runner_audit"
}

func NewRunner(params Params) *runnerMiddleware {
	return &runnerMiddleware{
		baseMiddleware: baseMiddleware{
			l:       params.L,
			db:      params.DB,
			context: "runner",
		},
	}
}
