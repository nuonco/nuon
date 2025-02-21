package api

import (
	"context"
	"fmt"
	"time"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/pkg/retry"
)

type Params struct {
	fx.In

	L   *zap.Logger `name:"dev"`
	Cfg *internal.Config
}

func New(params Params) (nuonrunner.Client, error) {
	retryer, err := retry.New(
		retry.WithMaxAttempts(5),
		retry.WithSleep(time.Second),
		retry.WithTimeout(time.Second*10),
		retry.WithCBHook(func(ctx context.Context, attempt int) error {
			l, err := pkgctx.Logger(ctx)
			if err != nil {
				// if not logger is found in the context, log with the default built in logger
				params.L.Warn("retrying request to runner-api", zap.Int("attempt", attempt))
				return nil
			}

			l.Warn("retrying request to runner-api", zap.Int("attempt", attempt))
			return nil
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get retryer")
	}

	api, err := nuonrunner.New(
		nuonrunner.WithURL(params.Cfg.RunnerAPIURL),
		nuonrunner.WithRunnerID(params.Cfg.RunnerID),
		nuonrunner.WithAuthToken(params.Cfg.RunnerAPIToken),
		nuonrunner.WithRetryer(retryer),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize runner: %w", err)
	}

	return api, nil
}
