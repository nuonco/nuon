package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/dev"
)

func (c *cli) registerRunLocal() error {
	if os.Getenv("ENV") != "development" {
		return nil
	}

	runCmd := &cobra.Command{
		Use:  "run-local",
		Long: "run-local runs the runner locally automatically using the admin api to fetch a runner-id and token, unless they are set.",
		Run:  c.runLocalRun,
	}

	rootCmd.AddCommand(runCmd)
	return nil
}

func (c *cli) runLocalRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("must pass in a valid runner-id or \"org\"|\"install\" to select the most current one")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	devver, err := dev.New(args[0])
	if err != nil {
		log.Fatalf("unable to initialize devver %s", err)
	}
	if err := devver.Init(ctx); err != nil {
		log.Fatalf("unable to initialize: %s", err)
	}

	fmt.Println("running runner like usual")
	arg := args[0]
	switch arg {
	case "org", "build":
		c.runBuild(cmd, nil)
	case "install":
		c.runInstall(cmd, nil)
	default:
		log.Fatalf("we know naught of this arg: %s", arg)

	}
}
