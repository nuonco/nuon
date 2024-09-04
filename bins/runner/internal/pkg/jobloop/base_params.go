package jobloop

import (
	"context"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type BaseParams struct {
	fx.In

	LC fx.Lifecycle

	Ctx         context.Context
	Client      nuonrunner.Client
	Settings    *settings.Settings
	L           *zap.Logger
	ErrRecorder *errs.Recorder
}
