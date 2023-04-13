package interceptors

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"go.uber.org/zap"
)

func LoggerInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		l := zap.L()

		resp, err := next(ctx, request)
		if err != nil {
			l.Error(fmt.Sprintf("error %s", err.Error()))
		}

		return resp, err
	})
}
