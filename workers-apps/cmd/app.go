package cmd

import (
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	shared "github.com/powertoolsdev/workers-apps/internal"
	"github.com/powertoolsdev/workers-apps/internal/provision"
	"github.com/powertoolsdev/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/workers-apps/internal/provision/repository"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var appsCmd = &cobra.Command{
	Use:   "domain",
	Short: "Run the app workers",
	Run:   appRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(appsCmd)
}

func appRun(cmd *cobra.Command, _ []string) {
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Base.Env {
	case config.Local, config.Development:
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

	l.Debug("starting domain workers", zap.Any("config", cfg))
	if err := runAppWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runAppWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	w := worker.New(c, "apps", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
	})

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid domain config: %w", err)
	}

	wkflow := provision.NewWorkflow(cfg)
	w.RegisterWorkflow(wkflow.Provision)
	w.RegisterActivity(provision.NewActivities())

	// register provision/project workflow
	pwkflow := project.NewWorkflow(cfg)
	w.RegisterWorkflow(pwkflow.ProvisionProject)
	pacts := project.NewActivities()
	w.RegisterActivity(pacts)

	// register provision/project workflow
	rwkflow := repository.NewWorkflow(cfg)
	w.RegisterWorkflow(rwkflow.ProvisionRepository)
	racts := repository.NewActivities()
	w.RegisterActivity(racts)

	if err := w.Run(interruptCh); err != nil {
		return err
	}
	return nil
}
