package org

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "orgs")

	// org defaults
	config.RegisterDefault("waypoint_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("waypoint_bootstrap_token_namespace", "default")

	// orgs k8s access defaults
	config.RegisterDefault("orgs_k8s_cluster_id", "orgs-stage-main")
	config.RegisterDefault("orgs_k8s_role_arn", "arn:aws:iam::766121324316:role/extra-auth-eks-workers-orgs-orgs-stage-main")
	config.RegisterDefault("orgs_k8s_ca_data", "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1UQXlPREU0TURNME5Gb1hEVE15TVRBeU5URTRNRE0wTkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTWlKCmc4ZEprQXlqb2JkeTdKQmhyQ2dpVzhOOTRnWW0wQVkrcks0aHJzZ1FBV1c5bUJmQm1xL05sZXpUSGY4Yng2WVYKa1gvdkhUU1I5QlRvOUpITGM0ME9EM05GaXpibGdTMFh6U3BPeE10TDBLeVFMbk5pVlBMRTZPU1dPN09uUjFOSApiWjF1T3M5VFNMU0trUkZHK21VVVZmZndaQ1YyTG81V0JWWFg4Q2JwaWhnRkU0U2NNc0dydmRjem9OVzlsMk16CkxrTFFJcE9GSmFXbHd3TlBmZnZzSXdJR1ZIaTBLd1lXaDFzbDM5azBUb3NNbXFaSW5oWWVabkUrYmg0NFh5WFMKOTVPbFFYbnpzd3U1TjY2MGJSQUJFVnJSOG1iNUd0Q3h4T1dXVjcrZHRiQU5jdEQvUlhlRUJqSzNyb3ZycXFDZQo2ZlJ1eWQ1RU1BUHgwQW5talprQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZPV0VmbC8va3c4Q2Z6NWo4SjVJalZkWkhPeTlNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBSldUaFBlS2pmOVZyQ2dlbnVDNgpLeUk4cXBzSXNxbG9JVEFueG00NENlYjR5bTJDY0hnQ2tNSEpCczhQdURHeGhldU9FOXJ5TTk4SU9SOUpFVHU0CmUwNWsvSUhQeFFGMWk5eldqcjIrREx3QjZnbGx5TGFQbEVCTXk2NTE4V1JpOStUM3ZnZGRiektyNnJTSnpISEsKajNMV0dJM1FQVVVqZEEvYXVoRElvVWdGcDJPOFFmRUI4UG96N0QrQVNmMkdyZEw4SlN2eTdsb0NxNW04RGs1VQpZRDNYY0JhRGMwd0F6T2xua1pyWWYzWWZFMXRwbzNGL3prbU9sOWRoemZhVWdZN0hOb0prOERSRjQ5aXRYclVsCjk2L1QvbzhWdjRHYUxqVHhkUUgyM0oxTzJoTExhRXdmOEQ2dXE3bG8rT2Z6L3NMTHNCUDNsRzFHcEhwRzBSOEgKVllzPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")
	config.RegisterDefault("orgs_k8s_public_endpoint", "https://891656020B6C382DE7E0D96E1B86D224.gr7.us-west-2.eks.amazonaws.com")

	// orgs buckets defaults
	config.RegisterDefault("org_deployments_bucket_name", "nuon-org-deployments-stage")
	config.RegisterDefault("org_installations_bucket_name", "nuon-org-installations-stage")
	config.RegisterDefault("org_orgs_bucket_name", "nuon-orgs-stage")
}

type Config struct {
	config.Base `config:",squash"`

	// NOTE: top level configuration for starting the worker
	BotsSlackWebhookURL string `config:"bots_slack_webhook_url"`
	TemporalHost        string `config:"temporal_host" validate:"required"`
	TemporalNamespace   string `config:"temporal_namespace" validate:"required"`

	WaypointServerRootDomain        string `config:"waypoint_server_root_domain" validate:"required" json:"waypoint_server_root_domain"`
	WaypointBootstrapTokenNamespace string `config:"waypoint_bootstrap_token_namespace" validate:"required" json:"waypoint_bootstrap_token_namespace"`

	// IAM roles used to manage things in orgs account
	OrgsIAMAccessRoleArn    string `config:"orgs_iam_access_role_arn" validate:"required"`
	OrgsBucketAccessRoleArn string `config:"orgs_bucket_access_role_arn" validate:"required"`
	OrgsKMSAccessRoleArn    string `config:"orgs_kms_access_role" validate:"required"`

	// configs needed to access the orgs cluster, which runs all org runner/servers
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id" json:"orgs_k8s_cluster_id" validate:"required"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn" json:"orgs_k8s_role_arn" validate:"required"`
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data" json:"orgs_k8s_ca_data" validate:"required"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint" json:"orgs_k8s_public_endpoint" validate:"required"`

	// configs needed for setting up permissions to org resources
	OrgInstallationsBucketName string `config:"org_installations_bucket_name" json:"org_installations_bucket_name" validate:"required"`
	OrgDeploymentsBucketName   string `config:"org_deployments_bucket_name" json:"org_deployments_bucket_name" validate:"required"`
	OrgKeyValuesBucketName     string `config:"org_key_values_bucket_name" json:"org_key_values_bucket_name" validate:"required"`
	OrgsBucketName             string `config:"org_orgs_bucket_name" json:"org_orgs_bucket_name" validate:"required"`

	// configs needed to grant the workers ability to assume org iam roles
	WorkersIAMRoleARNPrefix string `config:"workers_iam_role_arn_prefix" validate:"required"`
	SupportIAMRoleARN       string `config:"support_iam_role_arn" validate:"required"`

	// configs needed to create an IAM role for the ODR runner in the orgs account
	OrgsIAMOidcProviderURL string `config:"orgs_iam_oidc_provider_url" validate:"required"`
	OrgsIAMOidcProviderArn string `config:"orgs_iam_oidc_provider_arn" validate:"required"`
	OrgsECRRegistryArn     string `config:"orgs_ecr_registry_arn" validate:"required"`

	SandboxBucketARN string `config:"sandbox_bucket_arn" validate:"required"`
	SandboxKeyARN    string `config:"sandbox_key_arn" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
