package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerConsumer() error {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "run kafka consumers (heartbeats, logs, ...)",
		Run:   c.runConsumer,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

// runConsumer runs all Kafka consumers in a single process, decoupled from the
// Temporal workers. Infrastructure providers are lazy, so Temporal is never
// constructed here — only what the consumers depend on (config, ClickHouse,
// metrics, logging).
func (c *cli) runConsumer(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{}
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers, fxmodules.KafkaConsumersModule)

	fx.New(providers...).Run()
}
