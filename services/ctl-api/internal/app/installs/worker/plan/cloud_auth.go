package plan

import (
	"fmt"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
		azureAuth = &azurecredentials.Config{
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionID:       stackOutputs.AzureStackOutputs.SubscriptionID,
				SubscriptionTenantID: stackOutputs.AzureStackOutputs.SubscriptionTenantID,
			},
		}
		// Legacy installs have no per-operation identity; fall back to the runner's
		// ambient identity.
		if roleSelection.RoleARN != "" {
			azureAuth.ManagedIdentityClientID = roleSelection.RoleARN
		} else {
			azureAuth.UseDefault = true
		}
	case stackOutputs.GCPStackOutputs != nil:
		// An empty impersonation target silently falls back to the runner's
		// ambient identity. That is correct only for legacy stacks that predate
		// per-operation service accounts (no SA emails at all); on a stack that
		// does manage them, an empty selection means the role is disabled or
		// grants nothing, and running with ambient permissions instead would be
		// wrong. Mirrors the Azure legacy fallback above.
		gcpOut := stackOutputs.GCPStackOutputs
		stackManagesSAs := gcpOut.ProvisionSAEmail != "" || gcpOut.MaintenanceSAEmail != "" || gcpOut.DeprovisionSAEmail != ""
		if roleSelection.RoleARN == "" && stackManagesSAs {
			return nil, fmt.Errorf("unable to build cloud auth, missing role identifier")
		}

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
