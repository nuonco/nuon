package cmd

import (
	"log"

	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/spf13/cobra"
	tworker "go.temporal.io/sdk/worker"
)

var runWorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "run worker",
	Run:   runWorker,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runWorkerCmd)
}

func runWorker(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("no domain passed: please pass one of apps|orgs|deploys|builds etc")
	}
	workerDomain := args[0]

	app, err := newApp(cmd.Flags())
	if err != nil {
		log.Fatalf("unable to load server: %s", err)
	}
	sharedActs, err := sharedactivities.New(app.v,
		sharedactivities.WithTemporalHost(app.cfg.TemporalHost))
	if err != nil {
		log.Fatalf("error during worker setup: %s", err)
	}

	var initFn func(sa *sharedactivities.Activities) (worker.Worker, error)
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

	wkr, err := initFn(sharedActs)
	if err != nil {
		log.Fatalf("unable to initialize %s worker: %s", workerDomain, err)
	}

	interruptCh := tworker.InterruptCh()
	if err := wkr.Run(interruptCh); err != nil {
		log.Fatalf("worker exited: %s", err)
	}
}
