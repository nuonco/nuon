package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/management"
	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/health"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/log"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/shutdownbeacon"
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
		Long:  "Authenticate with a cloud provider using instance credentials and store the runner token.",
		Run:   c.runFetchToken,
	}
	fetchTokenCmd.Flags().Bool("json", false, "Output result as JSON (does not write token to disk)")
	fetchTokenCmd.Flags().String("platform", "", "Cloud platform to use for authentication (aws, azure, gcp). Defaults to auto-detect.")

	mngCmd.AddCommand(fetchTokenCmd)
	rootCmd.AddCommand(mngCmd)
	return nil
}

func (c *cli) runMng(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{fx.Provide(log.NewSystem)}
	providers = append(c.commonProviders(), providers...)
	providers = append(providers, management.GetJobs()...)
	providers = append(providers, fx.Provide(shutdownbeacon.New))
	// add mng and heartbeater to the mng process
	providers = append(providers,
		[]fx.Option{
			// provide process for the heartbeater
			fx.Supply(fx.Annotate("mng", fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			// NOTE: we do not include the `operations` job loops here
			// sandbox control API

			// start heartbeater, process registrar, and shutdown poller
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
			fx.Invoke(func(*process.Registrar) {}),
			fx.Invoke(func(*process.ShutdownPoller) {}),
			fx.Invoke(func(*shutdownbeacon.Beacon) {}),
			// serve /healthz for the Azure VMSS Application Health extension
			fx.Provide(health.New),
			fx.Invoke(func(*health.Server) {}),
		}...,
	)
	// run
	fx.New(providers...).Run()
}

func (c *cli) runFetchToken(cmd *cobra.Command, _ []string) {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")
	platform, _ := cmd.Flags().GetString("platform")

	// Fall back to env var if flag not set.
	if platform == "" {
		platform = os.Getenv("RUNNER_PLATFORM")
	}

	apiURL := os.Getenv("RUNNER_API_URL")
	if apiURL == "" {
		apiURL = "https://runner.nuon.co"
	}

	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(apiURL),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create API client: %v\n", err)
		os.Exit(1)
	}

	// Azure uses the TokenFetcher interface; AWS/GCP use existing inline code paths.
	var result *fetchtoken.FetchTokenResult
	if platform == "azure" {
		runnerID := os.Getenv("RUNNER_ID")
		if jsonOutput {
			result, err = fetchtoken.FetchTokenAzure(ctx, apiClient, runnerID)
		} else {
			result, err = fetchtoken.FetchAndStoreTokenAzure(ctx, apiClient, runnerID)
		}
	} else {
		authMethod := os.Getenv("RUNNER_AUTH_METHOD")
		runnerID := os.Getenv("RUNNER_ID")
		fmt.Fprintf(os.Stderr, "fetch-token: api_url=%s auth_method=%q runner_id=%s\n", apiURL, authMethod, runnerID)
		if jsonOutput {
			result, err = fetchtoken.FetchToken(ctx, apiClient, authMethod, runnerID)
		} else {
			result, err = fetchtoken.FetchAndStoreToken(ctx, apiClient, authMethod, runnerID)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch token: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode result: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("authentication successful\n")
	fmt.Printf("  runner_id:   %s\n", result.RunnerID)
	fmt.Printf("  instance_id: %s\n", result.InstanceID)
	if result.AccountID != "" {
		fmt.Printf("  account_id:  %s\n", result.AccountID)
	}
	if result.ProjectID != "" {
		fmt.Printf("  project_id:  %s\n", result.ProjectID)
	}
	fmt.Printf("  token_path:  %s\n", result.TokenPath)
}
