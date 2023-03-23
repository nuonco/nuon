package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/config"
	"github.com/powertoolsdev/mono/pkg/common/temporalzap"
	"github.com/powertoolsdev/mono/pkg/sender"
	shared "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup/iam"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup/kms"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup/runner"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup/server"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/teardown"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Run the org workers",
	Run:   orgRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(orgCmd)
}

func orgRun(cmd *cobra.Command, args []string) {
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %s", err))
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

	l.Debug("starting org workers", zap.Any("config", cfg))
	if err := runOrgWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runOrgWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	w := worker.New(c, "org", worker.Options{})

	var (
		n   signup.NotificationSender
		err error
	)

	l := zap.L()
	v := validator.New()

	// NOTE(jdt): this isn't my favorite
	switch cfg.Env {
	case config.Local, config.Development:
		l.Info("using noop notification sender")
		n = sender.NewNoopSender()
	default:
		n, err = sender.NewSlackSender(cfg.BotsSlackWebhookURL, l)
		if err != nil {
			l.Warn("failed to create slack notifier, using noop", zap.Error(err))
			n = sender.NewNoopSender()
		}
	}

	wkflow := signup.NewWorkflow(cfg)
	w.RegisterWorkflow(wkflow.Signup)
	w.RegisterActivity(signup.NewActivities(n))

	w.RegisterWorkflow(teardown.Teardown)
	w.RegisterActivity(teardown.NewActivities())

	runiFlow := runner.NewWorkflow(cfg)
	w.RegisterWorkflow(runiFlow.Install)
	w.RegisterActivity(runner.NewActivities(v, cfg))

	srvWkflow := server.NewWorkflow(cfg)
	w.RegisterWorkflow(srvWkflow.ProvisionServer)
	w.RegisterActivity(server.NewActivities(v))

	iamWkflow := iam.NewWorkflow(cfg)
	w.RegisterWorkflow(iamWkflow.ProvisionIAM)
	w.RegisterActivity(iam.NewActivities())

	kmsWkflow := kms.NewWorkflow(cfg)
	w.RegisterWorkflow(kmsWkflow.ProvisionKMS)
	w.RegisterActivity(kms.NewActivities())

	if err := w.Run(interruptCh); err != nil {
		return err
	}
	return nil
}
