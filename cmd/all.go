package cmd

import (
	"log"
	"sync"

	"github.com/powertoolsdev/go-common/temporalzap"
	"github.com/powertoolsdev/go-config/pkg/config"
	shared "github.com/powertoolsdev/workers-executors/internal"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var allCmd = &cobra.Command{
	Use:    "all",
	Short:  "Run all workers",
	Run:    runAll,
	PreRun: config.ConfigureService[*cobra.Command],
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(allCmd)
	flags := allCmd.Flags()

	flags.String("service_name", "workers-executors", "the name of the service")
	flags.String("temporal_host", "", "the temporal host and port")
	flags.String("temporal_namespace", "", "the temporal namespace")
}

type workerFn func(client.Client, *zap.Logger, shared.Config, <-chan interface{}) error

func runAll(cmd *cobra.Command, args []string) {
	l := zap.L()
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		l.Fatal("failed to load config", zap.Error(err))
	}

	if err := cfg.Validate(); err != nil {
		l.Fatal("failed to validate config", zap.Error(err))
	}

	c, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
		Logger:    temporalzap.NewLogger(l),
	})
	if err != nil {
		l.Fatal("failed to instantiate temporal client", zap.Error(err))
	}
	defer c.Close()

	ch := worker.InterruptCh()

	wg := new(sync.WaitGroup)

	workflows := []workerFn{
		runExecutorWorkers,
	}

	wg.Add(len(workflows))

	l.Debug("starting all workers", zap.Any("config", cfg))
	for _, worker := range workflows {
		go func(fn workerFn) {
			if err := fn(c, l, cfg, ch); err != nil {
				log.Fatalf("error in worker: %s", err)
			}
			wg.Done()
		}(worker)
	}

	wg.Wait()
}
