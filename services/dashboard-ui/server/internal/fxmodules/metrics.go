package fxmodules

import (
	"context"
	"fmt"
	"os"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/apiroutes"
)

func NewAPIRouteClassifier(lc fx.Lifecycle, cfg *internal.Config, l *zap.Logger) *apiroutes.Classifier {
	c := apiroutes.NewClassifier(cfg.APIUrl, l)

	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go c.Run(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})

	return c
}

func NewMetricsWriter(lc fx.Lifecycle, cfg *internal.Config, l *zap.Logger) (metrics.Writer, error) {
	v := validator.New()

	mw, err := metrics.New(v,
		metrics.WithDisable(cfg.DisableMetrics),
		metrics.WithTags(
			fmt.Sprintf("service_deployment:%s", cfg.ServiceDeployment),
			"service_type:dashboard_ui",
		),
		metrics.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create metrics writer: %w", err)
	}

	if !cfg.DisableMetrics {
		tracer.Start(
			tracer.WithRuntimeMetrics(),
			tracer.WithDogstatsdAddr(fmt.Sprintf("%s:8125", os.Getenv("HOST_IP"))),
		)

		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				tracer.Stop()
				return nil
			},
		})
	}

	return mw, nil
}
