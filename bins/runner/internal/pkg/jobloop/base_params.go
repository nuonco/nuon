package jobloop

import (
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/errs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/pkg/metrics"
)

type BaseParams struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner

	Client      nuonrunner.Client
	Settings    *settings.Settings
	Cfg         *internal.Config
	ErrRecorder *errs.Recorder
	MW          metrics.Writer

	L *zap.Logger `name:"system"`
}
