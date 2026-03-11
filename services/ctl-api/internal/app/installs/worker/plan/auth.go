package plan

import (
	"fmt"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CreatePlanAuth creates PlanAuth configuration from install stack outputs
// The roleARN parameter should be the selected role ARN based on operation type and configuration
func CreatePlanAuth(stackOutputs app.InstallStackOutputs, roleARN, sessionName string) (*plantypes.PlanAuth, error) {
	planAuth := &plantypes.PlanAuth{}

	switch {
	case stackOutputs.AWSStackOutputs != nil:
		if roleARN == "" {
			return nil, fmt.Errorf("role ARN is required for AWS stack")
		}

		planAuth.AWSAuth = &awscredentials.Config{
			Region: stackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: sessionName,
				RoleARN:     roleARN,
			},
		}

	case stackOutputs.AzureStackOutputs != nil:
		azureOutputs := stackOutputs.AzureStackOutputs
		planAuth.AzureAuth = &azurecredentials.Config{
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionID:       azureOutputs.SubscriptionID,
				SubscriptionTenantID: azureOutputs.SubscriptionTenantID,
			},
			UseDefault: true,
		}

	case stackOutputs.GCPStackOutputs != nil:
		planAuth.GCPAuth = &gcpcredentials.Config{
			ProjectID:                 stackOutputs.GCPStackOutputs.ProjectID,
			Region:                    stackOutputs.GCPStackOutputs.Region,
			ImpersonateServiceAccount: roleARN,
		}

	default:
		return nil, fmt.Errorf("no supported cloud provider found in stack outputs")
	}

	return planAuth, nil
}
