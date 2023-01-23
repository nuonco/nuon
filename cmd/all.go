package cmd

import (
	"fmt"
	"log"
	"sync"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	shared "github.com/powertoolsdev/workers-executors/internal"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all workers",
	Run:   runAll,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(allCmd)
}

type workerFn func(client.Client, shared.Config, <-chan interface{}) error

func runAll(cmd *cobra.Command, args []string) {
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate config: %s", err))
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Env {
	case config.Development:
		l, err = zap.NewDevelopment()
	default:
		l, err = zap.NewProduction()
	}
	zap.ReplaceGlobals(l)

	if err != nil {
		fmt.Printf("failed to instantiate logger: %v\n", err)
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
			if err := fn(c, cfg, ch); err != nil {
				log.Fatalf("error in worker: %s", err)
			}
			wg.Done()
		}(worker)
	}

	wg.Wait()
}
