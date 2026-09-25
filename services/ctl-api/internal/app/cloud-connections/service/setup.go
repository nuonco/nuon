package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SetupResponse struct {
	IssuerURL      string                          `json:"issuer_url"`
	Subject        string                          `json:"subject"`
	Audience       string                          `json:"audience"`
	TrustPolicy    map[string]any                  `json:"trust_policy"`
	Terraform      string                          `json:"terraform"`
	CLI            string                          `json:"cli"`
	CloudFormation string                          `json:"cloudformation"`
	Capabilities   []app.CloudConnectionCapability `json:"capabilities"`
	Repositories   []string                        `json:"repositories,omitempty"`
}

func subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

func (s *service) setup(connection *app.CloudConnection, repositories []string) SetupResponse {
	if s.issuer == nil {
		return SetupResponse{Subject: subject(connection), Audience: "sts.amazonaws.com", Capabilities: connection.Capabilities, Repositories: repositories}
	}
	issuerURL := s.issuer.Issuer()
	issuer, _ := url.Parse(issuerURL)
	conditionPrefix := strings.TrimPrefix(issuer.Host+issuer.Path, "/")
	trustPolicy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{map[string]any{
			"Effect":    "Allow",
			"Principal": map[string]string{"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", connection.TargetID, conditionPrefix)},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]any{"StringEquals": map[string]string{
				conditionPrefix + ":aud": "sts.amazonaws.com",
				conditionPrefix + ":sub": subject(connection),
			}},
		}},
	}
	policyJSON, _ := json.Marshal(trustPolicy)
	capabilitiesJSON, _ := json.Marshal(connection.Capabilities)
	repositoriesJSON, _ := json.Marshal(repositories)
	terraform := fmt.Sprintf(`module "nuon_cloud_connection" {
  source  = "nuonco/ecr-access/aws"
  nuon_issuer = %q
  nuon_subject = %q
  capabilities = %s
  repositories = %s
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON)
	cli := fmt.Sprintf("aws iam create-open-id-connect-provider --url %s --client-id-list sts.amazonaws.com --thumbprint-list <thumbprint>\naws iam create-role --role-name <role-name> --assume-role-policy-document '%s'", issuerURL, policyJSON)
	cloudFormation := fmt.Sprintf("NuonConnectionRole:\n  Type: AWS::IAM::Role\n  Properties:\n    AssumeRolePolicyDocument: %s\n", string(policyJSON))
	return SetupResponse{IssuerURL: issuerURL, Subject: subject(connection), Audience: "sts.amazonaws.com", TrustPolicy: trustPolicy, Terraform: terraform, CLI: cli, CloudFormation: cloudFormation, Capabilities: connection.Capabilities, Repositories: repositories}
}
