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

func (c *cli) runConsumer(cmd *cobra.Command, _ []string) error {
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
