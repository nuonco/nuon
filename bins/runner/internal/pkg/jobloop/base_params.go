package jobloop

import (
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

type BaseParams struct {
	fx.In

	LC fx.Lifecycle

	Client      nuonrunner.Client
	Settings    *settings.Settings
	Cfg         *internal.Config
	ErrRecorder *errs.Recorder
	MW          metrics.Writer

	L *zap.Logger `name:"system"`
}
