package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/pkg/api"
)

const (
	adminAPIURL string = "http://localhost:8082"
)

func (c *cli) registerRunLocal() error {
	if os.Getenv("RUNNER_ENABLE_LOCAL") != "true" {
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

func (c *cli) runLocalRun(cmd *cobra.Command, _ []string) {
	runnerID := os.Getenv("RUNNER_ID")
	runnerAPIToken := os.Getenv("RUNNER_API_TOKEN")

	v := validator.New()
	ctx := context.Background()
	apiClient, err := api.New(v, api.WithURL(adminAPIURL))
	if err != nil {
		log.Fatal("unable to create admin api url for run-local")
	}

	if runnerID == "" {
		fmt.Println("no runner_id set, looking up a runner from the api")
		runners, err := apiClient.ListRunners(ctx)
		if err != nil {
			log.Fatalf("unable to list runners from api: %s", err)
		}

		if len(runners) < 1 {
			log.Fatalf("no runners found locally")
		}

		fmt.Println("setting runner ID to ", runners[0].ID)
		runnerID = runners[0].ID
		os.Setenv("RUNNER_ID", runnerID)
	}

	if runnerAPIToken == "" {
		fmt.Println("no runner_api_token set, looking up a token from the api")
		token, err := apiClient.GetRunnerServiceAccountToken(ctx, runnerID, time.Hour)
		if err != nil {
			log.Fatalf("unable to get token from api: %s", err)
		}

		fmt.Println("setting runner api token to ", token)
		os.Setenv("RUNNER_API_TOKEN", token)
	}

	fmt.Println("running runner like usual")
	c.runRun(cmd, nil)
}
