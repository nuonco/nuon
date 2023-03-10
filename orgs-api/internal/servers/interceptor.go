package servers

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
)

func EnsureShortIDInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		err := EnsureShortID(request.Any())
		if err != nil {
			return nil, fmt.Errorf("unable to ensure shortIDs: %w", err)
		}

		return next(ctx, request)
	})
}
