package dev

import (
	"context"
	"fmt"
	"os"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/pkg/errors"

	awsassumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/retry"
)

// if the runner is not in sandbox mode, and has an IAM role ARN, we assume that and set it in the environment,
// so we can mimic the IAM role of an install or org.
func (d *devver) initCreds(ctx context.Context) error {
	api, err := nuonrunner.New(
		nuonrunner.WithURL(os.Getenv("RUNNER_API_URL")),
		nuonrunner.WithRunnerID(d.runnerID),
		nuonrunner.WithAuthToken(d.runnerAPIToken),
	)
	if err != nil {
		return errors.Wrap(err, "unable to initialize runner api client")
	}

	settings, err := api.GetSettings(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get settings")
	}

	if settings.SandboxMode {
		fmt.Println("runner is in sandbox mode, skipping setting credentials")
		return nil
	}

	if settings.LocalAwsIamRoleArn == "" {
		fmt.Println("runner group has no local IAM role set, ignoring")
		return nil
	}

	fmt.Println("fetching credentials for " + settings.LocalAwsIamRoleArn)

	fn := func(ctx context.Context) error {
		assumer, err := awsassumerole.New(d.v,
			awsassumerole.WithRoleARN(settings.LocalAwsIamRoleArn),
			awsassumerole.WithRoleSessionName("nuon-ctl"),
		)
		if err != nil {
			return errors.Wrap(err, "unable to get role assumer")
		}

		cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to assume role")
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to fetch credentials")
		}

		os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
		return nil
	}

	if err := retry.Retry(ctx, fn, retry.WithMaxAttempts(10), retry.WithSleep(5)); err != nil {
		return err
	}

	fmt.Println("successfully set credentials in environment")
	return nil
}
