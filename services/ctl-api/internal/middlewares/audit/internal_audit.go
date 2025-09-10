package audit

type internalMiddleware struct {
	*baseMiddleware
}

func (m *internalMiddleware) Name() string {
	return "internal_audit"
}

func NewInternal(params Params) *internalMiddleware {
	return &internalMiddleware{
		baseMiddleware: newBaseMiddleware(params, "internal"),
	}
}
