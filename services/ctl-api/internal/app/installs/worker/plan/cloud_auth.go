package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

// CloudAuth is a wrapper around multiple auth configurations, its not a construct by itself, mostly used for passing
// around auth information to and from function calls
type CloudAuth struct {
	AWS   *awscredentials.Config
	Azure *azurecredentials.Config
	GCP   *gcpcredentials.Config
}

func getCloudAuth(
	roleSelection *operationroles.RoleSelection,
	stackOutputs *app.InstallStackOutputs,
	sessionName string,
) (*CloudAuth, error) {
	var awsAuth *awscredentials.Config
	var azureAuth *azurecredentials.Config
	var gcpAuth *gcpcredentials.Config
	switch {
	case stackOutputs.AWSStackOutputs != nil:
		if roleSelection.RoleARN == "" {
			return nil, fmt.Errorf("unable to build cloud auth, missing role identifier")
		}

		awsAuth = &awscredentials.Config{
			Region: stackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: sessionName,
				RoleARN:     roleSelection.RoleARN,
			},
		}

	case stackOutputs.AzureStackOutputs != nil:
		// currently azure does not support role based access control and only supports
		// single SubscriptionID based auth
		// once that is fixed we should update below impleentation and operationroles package as well to utilize roles
		azureAuth = &azurecredentials.Config{
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionID:       stackOutputs.AzureStackOutputs.SubscriptionID,
				SubscriptionTenantID: stackOutputs.AzureStackOutputs.SubscriptionTenantID,
			},
			UseDefault: true,
		}
	case stackOutputs.GCPStackOutputs != nil:
		// gcp uses default instance auth, no config needed
		gcpAuth = &gcpcredentials.Config{
			ProjectID:                 stackOutputs.GCPStackOutputs.ProjectID,
			Region:                    stackOutputs.GCPStackOutputs.Region,
			ImpersonateServiceAccount: roleSelection.RoleARN,
		}

	default:
		return nil, fmt.Errorf("no supported cloud provider found in stack outputs")
	}

	return &CloudAuth{
		Azure: azureAuth,
		AWS:   awsAuth,
		GCP:   gcpAuth,
	}, nil
}

// gcpRunnerInstanceRoleFallback reports whether the org opted back into the
// legacy GCP auth behavior where GKE cluster auth, GAR registry access, and
// action workflow env use the runner's instance service account instead of
// impersonating the selected operation-role SA. Any error resolving the
// feature keeps the default (impersonated) behavior.
func (p *Planner) gcpRunnerInstanceRoleFallback(ctx workflow.Context, stackOutputs *app.InstallStackOutputs) bool {
	if stackOutputs.GCPStackOutputs == nil {
		return false
	}

	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureGCPRunnerInstanceRole))
	if err != nil {
		if l, lerr := log.WorkflowLogger(ctx); lerr == nil {
			l.Warn("unable to check gcp-runner-instance-role feature; using impersonated auth", zap.Error(err))
		}
		return false
	}

	return enabled
}

// withoutGCPImpersonation returns a copy of the GCP auth config with the
// impersonation target cleared, so the runner falls back to its instance
// service account. The input is copied because the same *gcpcredentials.Config
// is shared with plan fields (Terraform/Pulumi env) that must keep
// impersonating.
func withoutGCPImpersonation(cfg *gcpcredentials.Config) *gcpcredentials.Config {
	if cfg == nil || cfg.ImpersonateServiceAccount == "" {
		return cfg
	}

	copied := *cfg
	copied.ImpersonateServiceAccount = ""
	return &copied
}
