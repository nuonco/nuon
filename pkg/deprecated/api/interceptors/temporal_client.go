package interceptors

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
)

func NewTemporalClientInterceptor(client temporal.Client) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, temporal.ContextKey{}, client)
			return next(ctx, request)
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
