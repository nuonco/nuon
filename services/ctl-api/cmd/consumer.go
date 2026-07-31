package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
	"github.com/nuonco/nuon/services/ctl-api/internal/health"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
)

var consumerName string

func (c *cli) registerConsumer() error {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "run kafka consumers (heartbeats, otel-logs, ...)",
		RunE:  c.runConsumer,
	}
	cmd.Flags().StringVar(&consumerName, "name", consumer.NameAll,
		fmt.Sprintf("which consumer to run: %s, or %q to run them all in one process", consumer.Names(), consumer.NameAll))
	rootCmd.AddCommand(cmd)
	return nil
}

// runConsumer runs the selected Kafka consumer, decoupled from the Temporal
// workers. Infrastructure providers are lazy, so Temporal is never constructed
// here — only what the consumers depend on (config, ClickHouse, metrics,
// logging).
//
// Deployed, a pod runs one consumer (`--name=heartbeats`) or a group of them
// that can share resources and a restart (`--name=otel-logs,otel-traces`); each
// still gets its own topic, consumer group and client. Locally `--name=all` runs
// them together.
func (c *cli) runConsumer(cmd *cobra.Command, _ []string) error {
	// Parsed before the fx graph is built so a bad --name is an immediate,
	// legible error rather than a provider failure buried in an fx trace.
	selection, err := consumer.NewSelection(consumerName)
	if err != nil {
		return err
	}

	providers := []fx.Option{}
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers, fx.Supply(selection))
	providers = append(providers, fxmodules.KafkaConsumersModule)

	providers = append(providers,
		fx.Provide(health.NewConsumerHealthcheck),
		fx.Invoke(func(lc fx.Lifecycle, hc *health.ConsumerHealthcheckServer) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					return hc.Start()
				},
				OnStop: func(ctx context.Context) error {
					return hc.Stop(ctx)
				},
			})
		}),
	)

	fx.New(providers...).Run()

	return nil
}
