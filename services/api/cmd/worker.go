package cmd

import (
	"log"

	"github.com/spf13/cobra"
	worker "go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"
)

var runWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "run worker",
	Run:   runWorker,
}
var workerDomain string

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runWorkerCmd)

	flags := runWorkerCmd.Flags()
	flags.String("domain", "", "which domain to run, or all for all")
	runWorkerCmd.PersistentFlags().StringVar(&workerDomain, "domain", "all", "worker domain")
}

func runWorker(cmd *cobra.Command, _ []string) {
	app, err := newApp(cmd.Flags())
	if err != nil {
		log.Fatalf("unable to load server: %s", err)
	}

	workers, err := app.loadWorkers(workerDomain)
	if err != nil {
		log.Fatalf("unable to load workers: %s", err)
	}

	interruptCh := worker.InterruptCh()
	g := new(errgroup.Group)
	for _, wkr := range workers {
		wkr := wkr
		g.Go(func() error {
			return wkr.Run(interruptCh)
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatalf("unable to wait for workers: %s", err)
	}
}
