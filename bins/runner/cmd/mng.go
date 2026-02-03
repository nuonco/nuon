package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/management"
	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/log"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerMng() error {
	mngCmd := &cobra.Command{
		Use:     "mng",
		Short:   "Run in management mode.",
		Long:    "Run in management mode and oversee an install mode process in a standalone VM.",
		Aliases: []string{"management"},
		Run:     c.runMng,
	}

	fetchTokenCmd := &cobra.Command{
		Use:   "fetch-token",
		Short: "Fetch and store the runner authentication token.",
		Long:  "Authenticate with AWS using instance credentials and store the runner token.",
		Run:   c.runFetchToken,
	}

	mngCmd.AddCommand(fetchTokenCmd)
	rootCmd.AddCommand(mngCmd)
	return nil
}

func (c *cli) runMng(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{fx.Provide(log.NewSystem)}
	providers = append(c.commonProviders(), providers...)
	providers = append(providers, management.GetJobs()...)
	// add mng and heartbeater to the mng process
	providers = append(providers,
		[]fx.Option{
			// provide process for the heartbeater
			fx.Supply(fx.Annotate("mng", fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			// NOTE: we do not include the `operations` job loops here
			// start registry and heartbeater
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
		}...,
	)
	// run
	fx.New(providers...).Run()
}

func (c *cli) runFetchToken(cmd *cobra.Command, _ []string) {
	ctx := context.Background()

	apiURL := os.Getenv("RUNNER_API_URL")
	if apiURL == "" {
		apiURL = "https://api.nuon.co"
	}

	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(apiURL),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create API client: %v\n", err)
		os.Exit(1)
	}

	result, err := fetchtoken.FetchAndStoreToken(ctx, apiClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("authentication successful\n")
	fmt.Printf("  runner_id:   %s\n", result.RunnerID)
	fmt.Printf("  instance_id: %s\n", result.InstanceID)
	fmt.Printf("  account_id:  %s\n", result.AccountID)
	fmt.Printf("  token_path:  %s\n", result.TokenPath)
}
