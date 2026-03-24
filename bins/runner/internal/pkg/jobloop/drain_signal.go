package jobloop

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

const drainTimeout = 60 * time.Minute

type DrainSignalParams struct {
	fx.In

	JobLoops       []JobLoop `group:"job_loops"`
	OperationsLoop []JobLoop `group:"operations"`
	Shutdowner     fx.Shutdowner
	L              *zap.Logger `name:"system"`
}

func RegisterDrainSignalHandler(params DrainSignalParams) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)

	go func() {
		<-sigCh
		params.L.Info("received SIGUSR1, draining all job loops")

		allLoops := append(params.JobLoops, params.OperationsLoop...)
		for _, loop := range allLoops {
			loop.Drain(drainTimeout)
		}

		params.L.Info("all job loops drained, shutting down")
		params.Shutdowner.Shutdown(fx.ExitCode(0))
	}()
}
