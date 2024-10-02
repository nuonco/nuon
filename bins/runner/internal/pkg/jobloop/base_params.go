package jobloop

import (
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type BaseParams struct {
	fx.In

	LC fx.Lifecycle

	Client      nuonrunner.Client
	Settings    *settings.Settings
	L           *zap.Logger
	ErrRecorder *errs.Recorder
}
