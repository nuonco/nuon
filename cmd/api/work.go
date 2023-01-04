package cmd

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "do some work",
	Run:   work,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(workCmd)
}

func work(cmd *cobra.Command, args []string) {
	var cfg Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
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

	// just hang out forever. replace with your logic
	for {
		l.Info("sleeping for 1 minute...")
		time.Sleep(1 * time.Minute)
	}
}
