package dev

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	awsassumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/retry"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
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

	// Azure RBAC setup must happen regardless of sandbox mode, because
	// local dev needs the AKS RBAC Cluster Admin role assigned to the
	// developer's identity (production VMs get it via the Bicep template).
	isAzure := settings.Platform == "azure-aks" || settings.Platform == "azure" || settings.Platform == "azure-acs"
	if isAzure {
		if err := d.initAzureCreds(ctx, settings); err != nil {
			return err
		}
		// Don't return early: the runner also needs AWS credentials for
		// pulling build artifacts from the management ECR (which is always
		// on AWS). Fall through to ensure AWS creds are available.
	}

	if settings.SandboxMode {
		fmt.Println("runner is in sandbox mode, skipping setting credentials")
		return nil
	}

	// For AWS installs, assume the install-specific IAM role.
	if settings.LocalAwsIamRoleArn != "" {
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

	// For GCP installs, the runner still needs AWS credentials to pull
	// artifacts from the management ECR. Azure installs use ACR natively
	// when the management registry is ACR.
	if isAzure {
		fmt.Println("Azure platform detected, management registry access uses Azure credentials")
		return nil
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" {
		fmt.Println("AWS credentials available via environment for management ECR access")
		return nil
	}

	fmt.Println("WARNING: no AWS credentials found in environment (AWS_PROFILE or AWS_ACCESS_KEY_ID). " +
		"The runner needs AWS credentials to pull build artifacts from the management ECR. " +
		"Run `aws sso login` and set AWS_PROFILE.")
	return nil
}

// initAzureCreds ensures the local az-login user has the AKS RBAC Cluster Admin role
// on the install's resource group. In production, the VMSS managed identity gets this
// role via the Bicep template; locally we need to assign it to the developer's identity.
func (d *devver) initAzureCreds(ctx context.Context, settings *models.AppRunnerGroupSettings) error {
	fmt.Println("azure platform detected, ensuring local AKS RBAC permissions")

	runner, err := d.apiClient.GetRunner(ctx, d.runnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	runnerGroup, err := d.apiClient.GetRunnerGroup(ctx, runner.RunnerGroupID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner group")
	}

	if runnerGroup.Type != "install" {
		fmt.Println("not an install runner, skipping Azure RBAC setup")
		return nil
	}

	install, err := d.apiClient.GetInstall(ctx, runnerGroup.OwnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	if install.AzureAccount == nil {
		return fmt.Errorf("install %s has no azure account configured", install.Id)
	}

	subscriptionID := install.AzureAccount.SubscriptionID
	if subscriptionID == "" {
		return fmt.Errorf("install %s has no subscription ID", install.Id)
	}

	// Scope to the subscription rather than a specific resource group, because
	// the AKS cluster may live in a different install's resource group (the
	// sandbox install) rather than this install's resource group.
	scope := fmt.Sprintf("/subscriptions/%s", subscriptionID)

	// Get the current az login user's object ID
	userOut, err := exec.CommandContext(ctx, "az", "ad", "signed-in-user", "show", "--query", "id", "-o", "tsv").CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to get current Azure user (is `az login` done?): %w: %s", err, string(userOut))
	}
	userObjectID := strings.TrimSpace(string(userOut))

	// Assign Azure Kubernetes Service RBAC Cluster Admin role
	fmt.Printf("assigning AKS RBAC Cluster Admin to user %s on %s\n", userObjectID, scope)
	out, err := exec.CommandContext(ctx, "az", "role", "assignment", "create",
		"--assignee-object-id", userObjectID,
		"--assignee-principal-type", "User",
		"--role", "Azure Kubernetes Service RBAC Cluster Admin",
		"--scope", scope,
	).CombinedOutput()
	if err != nil {
		// If the assignment already exists, az cli returns success, so this is a real error
		return fmt.Errorf("unable to assign AKS RBAC role: %w: %s", err, string(out))
	}

	fmt.Println("successfully ensured AKS RBAC Cluster Admin role for local dev")
	return nil
}
