package org

import "github.com/go-playground/validator/v10"

type Config struct {
	WaypointServerRootDomain        string `config:"waypoint_server_root_domain" validate:"required" json:"waypoint_server_root_domain"`
	WaypointBootstrapTokenNamespace string `config:"waypoint_bootstrap_token_namespace" validate:"required" json:"waypoint_bootstrap_token_namespace"`

	Bucket       string `config:"bucket" validate:"required" json:"bucket"`
	BucketRegion string `config:"bucket_region" validate:"required" json:"bucket_region"`

	// RoleArn is the role that we add to the sandbox EKS allowed roles so we can do other operations against it
	RoleArn string `config:"role_arn" validate:"required" json:"role_arn"`

	// configs needed to create an IAM role / ODR runner
	OrgsIAMAccessRoleArn         string `config:"orgs_iam_access_role_arn" validate:"required"`
	OrgsIAMOidcProviderURL       string `config:"orgs_iam_oidc_provider_url" validate:"required"`
	OrgsECRRegistryID            string `config:"orgs_ecr_registry_id" validate:"required"`
	OrgsIAMOidcFederationRoleArn string `config:"orgs_iam_oidc_federation_role_arn" validate:"required"`

	// configs needed to access the orgs cluster, which runs all org runner/servers
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id" json:"orgs_k8s_cluster_id"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn" json:"orgs_k8s_role_arn"`
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data" json:"orgs_k8s_ca_data"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint" json:"orgs_k8s_public_endpoint"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
