package metrics

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/version"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

type Params struct {
	fx.In

	Logger   *zap.Logger `name:"system"`
	Settings *settings.Settings
	V        *validator.Validate
	Cfg      *internal.Config
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
