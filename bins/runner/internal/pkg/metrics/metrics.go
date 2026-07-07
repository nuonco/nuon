package metrics

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"
)

type Params struct {
	fx.In

	Logger   *zap.Logger `name:"system"`
	Settings *settings.Settings
	V        *validator.Validate
	Cfg      *runnerconfig.Config
}

func New(params Params) (metrics.Writer, error) {
	tags := metrics.ToTags(params.Settings.Metadata, "version", version.Version, "git_ref", params.Cfg.GitRef, "service", "runner")

	disableMetrics := !params.Settings.EnableMetrics
	if os.Getenv("ENV") == "development" {
		disableMetrics = true
	}

	mw, err := metrics.New(params.V,
		metrics.WithDisable(disableMetrics),
		metrics.WithTags(tags...),
		metrics.WithLogger(params.Logger),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new metrics writer: %w", err)
	}

	return mw, nil
}
