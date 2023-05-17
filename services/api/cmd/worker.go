package cmd

import (
	"log"

	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/spf13/cobra"
	tworker "go.temporal.io/sdk/worker"
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

	var initFn func() (worker.Worker, error)
	switch workerDomain {
	case "apps":
		initFn = app.appsWorker
	case "builds":
		initFn = app.buildsWorker
	case "installs":
		initFn = app.installsWorker
	case "orgs":
		initFn = app.orgsWorker
	case "deploys":
		initFn = app.deploysWorker
	default:
		log.Fatalf("unknown domain: %s", workerDomain)
	}

	wkr, err := initFn()
	if err != nil {
		log.Fatalf("unable to initialize %s worker: %s", workerDomain, err)
	}

	interruptCh := tworker.InterruptCh()
	if err := wkr.Run(interruptCh); err != nil {
		log.Fatalf("worker exited: %s", err)
	}
}
