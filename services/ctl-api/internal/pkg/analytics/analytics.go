package analytics

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	V   *validator.Validate
	L   *zap.Logger
	Cfg *internal.Config
}

func New(params Params) (analytics.Writer, error) {
	w, err := analytics.New(params.V,
		analytics.WithDisable(params.Cfg.DisableAnalytics),
		analytics.WithSegmentKey(params.Cfg.SegmentWriteKey),
		analytics.WithLogger(params.L),
		analytics.WithGroupFn(groupFn),
		analytics.WithIdentifyFn(identifyFn),
		analytics.WithUserIDFn(userIDFn),
		analytics.WithProperties(map[string]interface{}{
			"platform": "ctl-api",
			"env":      params.Cfg.Env,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get analytics writer")
	}

	return w, nil
}
