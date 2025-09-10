package audit

type publicMiddleware struct {
	baseMiddleware
}

func (m *publicMiddleware) Name() string {
	return "public_audit"
}

func NewPublic(params Params) *publicMiddleware {
	return &publicMiddleware{
		baseMiddleware: baseMiddleware{
			l:       params.L,
			db:      params.DB,
			context: "public",
		},
	}
}
