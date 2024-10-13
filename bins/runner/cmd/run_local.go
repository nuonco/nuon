package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/spf13/cobra"

	smithytime "github.com/aws/smithy-go/time"

	awsassumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"

	"github.com/powertoolsdev/mono/pkg/api"
)

const (
	localFindRunnerPeriod time.Duration = time.Second
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
	runnerAPIToken := os.Getenv("RUNNER_API_TOKEN")
	adminAPIURL := os.Getenv("INTERNAL_API_URL")

	if len(args) < 1 {
		log.Fatal("must pass in a valid runner-id or orgs|installs to select the most current one")
	}

	v := validator.New()
	ctx := context.Background()
	apiClient, err := api.New(v, api.WithURL(adminAPIURL))
	if err != nil {
		log.Fatal("unable to create admin api url for run-local")
	}

	// fetch the correct runner ID to execute with
	var runnerID string
	switch args[0] {
	case "orgs", "installs":
		for runnerID == "" {
			fmt.Println("no runner_id set, looking up a runner from the api using type ", args[0])
			runners, err := apiClient.ListRunners(ctx, args[0])
			if err != nil {
				fmt.Println("unable to reach api, waiting for it to come online")
				smithytime.SleepWithContext(ctx, time.Second)
				continue
			}

			fmt.Println("no runner found, waiting for one to appear")
			if len(runners) < 1 {
				smithytime.SleepWithContext(ctx, time.Second)
				continue
			}

			fmt.Println("setting runner ID to ", runners[0].ID)
			runnerID = runners[0].ID
		}
	default:
		fmt.Printf("treating first argument as a runner id " + args[0] + " " + args[1])
		runnerID = args[0]
	}
	if runnerID == "" {
		log.Fatalf("no runner id")
	}
	fmt.Println("using runner id", runnerID)
	os.Setenv("RUNNER_ID", runnerID)

	// if the runner api token is not set, fetch one
	if runnerAPIToken == "" {
		for runnerAPIToken == "" {
			fmt.Println("no runner_api_token set, looking up a token from the api")
			token, err := apiClient.GetRunnerServiceAccountToken(ctx, runnerID, time.Hour)
			if err != nil {
				fmt.Println("waiting for runner to finish provisioning on ctl api")
				smithytime.SleepWithContext(ctx, time.Second)
				continue
			}

			fmt.Println("setting runner api token to ", token)
			runnerAPIToken = token
		}
	}
	os.Setenv("RUNNER_API_TOKEN", runnerAPIToken)

	// if the runner is not in sandbox mode, and has an IAM role ARN, we assume that and set it in the environment,
	// so we can mimic the IAM role of an install or org.
	api, err := nuonrunner.New(
		nuonrunner.WithURL(os.Getenv("RUNNER_API_URL")),
		nuonrunner.WithRunnerID(os.Getenv("RUNNER_ID")),
		nuonrunner.WithAuthToken(os.Getenv("RUNNER_API_TOKEN")),
	)
	if err != nil {
		log.Fatalf("unable to get api client: %s", err)
	}
	settings, err := api.GetSettings(ctx)
	if err != nil {
		log.Fatalf("unable to get settings: %s", err)
	}

	// this only works if the install role or org iam role grants access to the support role. This should be all of
	// our internal test IAM roles by default, and the infra-orgs managed roles
	if !settings.SandboxMode && settings.AwsIamRoleArn != "" {
		fmt.Println("fetching credentials for role to do real AWS operations " + settings.AwsIamRoleArn)
		assumer, err := awsassumerole.New(v,
			awsassumerole.WithRoleARN(settings.AwsIamRoleArn),
			awsassumerole.WithRoleSessionName("nuon-ctl"),
		)
		if err != nil {
			log.Fatalf("unable to get role assumer: %s", err)
		}

		cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
		if err != nil {
			log.Fatalf("unable to assume role: %s", err)
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			log.Fatalf("unable to fetch credentials: %s", err)
		}

		os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
	}

	fmt.Println("running runner like usual")
	c.runRun(cmd, nil)
}
