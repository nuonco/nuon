package kube

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
)

type ClusterInfo struct {
	// ID is the ID of the EKS cluster
	ID string `json:"id" hcl:"id" features:"template"`
	// Endpoint is the URL of the k8s api server
	Endpoint string `json:"endpoint" hcl:"endpoint" features:"template"`
	// CAData is the base64 encoded public certificate
	CAData string `json:"ca_data" hcl:"ca_data" features:"template"`

	EnvVars map[string]string `json:"env_vars" hcl:"env_vars" features:"template"`

	// KubeConfig will override the kube config, and be parsed instead of generating a new one
	KubeConfig string `json:"kube_config" faker:"-" hcl:"kube_config"`

	// If either an AWS auth or Azure auth is passed in, we will automatically use it to resolve credentials and set
	// them in the environment.
	//
	// NOTE: keep this comment detached from the field below (blank line above the
	// field). A doc comment on a $ref field makes go-swagger wrap the property in
	// allOf and generate it as an inline struct VALUE instead of a pointer — a
	// null aws_auth then round-trips to {} in the runner SDK and breaks kube auth
	// on non-AWS installs.

	AWSAuth   *awscredentials.Config   `json:"aws_auth" hcl:"aws_auth,block"`
	AzureAuth *azurecredentials.Config `json:"azure_auth" hcl:"azure_auth,block"`
	GCPAuth   *gcpcredentials.Config   `json:"gcp_auth" hcl:"gcp_auth,block"`

	// If this is set, we will _not_ use aws-iam-authenticator, but rather inline create the token
	Inline bool `json:"inline"`

	// TrustedRoleARN is the arn of the role that should be assumed to interact with the cluster
	// NOTE(JM): we are deprecating this
	TrustedRoleARN string `json:"trusted_role_arn" hcl:"trusted_role_arn"`
}

func (c *ClusterInfo) WithAWSAuth(auth *awscredentials.Config) {
	c.AWSAuth = auth
}

func (c *ClusterInfo) WithAzureAuth(auth *azurecredentials.Config) {
	c.AzureAuth = auth
}

func (c *ClusterInfo) WithGCPAuth(auth *gcpcredentials.Config) {
	c.GCPAuth = auth
}
