package interceptors

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

func MetricsInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			v := validator.New()

			writer, err := metrics.New(v, metrics.WithDisable(false))
			if err != nil {
				return nil, fmt.Errorf("unable to get metrics writer: %v", err)
			}

			startTS := time.Now()
			status := "ok"
			resp, err := next(ctx, request)
			if err != nil {
				status = "error"
			}

			tags := []string{
				"status:" + status,
				"endpoint:" + request.Spec().Procedure,
			}
			if os.Getenv("SERVICE_NAME") != "" {
				tags = append(tags, "service:"+os.Getenv("SERVICE_NAME"))
			}
			if os.Getenv("GIT_REF") != "" {
				tags = append(tags, "git_ref:"+os.Getenv("GIT_REF"))
			}

			writer.Incr("api.request.status", 1, tags)
			writer.Timing("api.request.latency", time.Since(startTS), tags)
			return resp, err
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
