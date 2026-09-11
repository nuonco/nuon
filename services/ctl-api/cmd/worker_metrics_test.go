package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func TestGeneralWorkerProviderGraph(t *testing.T) {
	providers := (&cli{}).providers()
	providers = append(providers,
		fx.NopLogger,
		fxmodules.WorkerInterceptorsModule,
		fxmodules.SharedWorkflowsModule,
		fxmodules.SlackLibsModule,
		fxmodules.GeneralWorkerModule,
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {})),
	)

	require.NoError(t, fx.ValidateApp(providers...))
}
