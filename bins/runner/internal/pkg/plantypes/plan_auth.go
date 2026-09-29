package plantypes

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
)

type PlanAuth struct {
	AWSAuth   *awscredentials.Config   `json:"aws_auth,omitempty"`
	AzureAuth *azurecredentials.Config `json:"azure_auth,omitempty"`
	GCPAuth   *gcpcredentials.Config   `json:"gcp_auth,omitempty`
}
