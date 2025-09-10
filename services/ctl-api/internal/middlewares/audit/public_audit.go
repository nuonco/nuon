package audit

type publicMiddleware struct {
	*baseMiddleware
}

func (m *publicMiddleware) Name() string {
	return "public_audit"
}

func NewPublic(params Params) *publicMiddleware {
	return &publicMiddleware{
		baseMiddleware: newBaseMiddleware(params, "public"),
	}
}
