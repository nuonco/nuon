package iam

import (
	"encoding/json"
	"fmt"

	workers "github.com/powertoolsdev/workers-orgs/internal"
)

func odrIAMName(orgID string) string {
	return fmt.Sprintf("org-odr-%s", orgID)
}

func runnerOdrServiceAccountName(orgID string) string {
	return fmt.Sprintf("waypoint-odr-%s", orgID)
}

func odrIAMPolicy(cfg workers.Config, orgID string) ([]byte, error) {
	policy := iamRolePolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleStatement{
			{
				Effect: "Allow",
				Action: []string{
					"ecr:*",
				},
				Resource: fmt.Sprintf("%s/%s/*", cfg.OrgsECRRegistryArn, orgID),
			},
			{
				Effect: "Allow",
				Action: []string{
					"ecr:GetAuthorizationToken",
				},
				Resource: "*",
			},
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("unable to convert policy to json: %w", err)
	}
	return byts, nil
}

func odrIAMTrustPolicy(cfg workers.Config, orgID string) ([]byte, error) {
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: "sts:AssumeRoleWithWebIdentity",
				Effect: "Allow",
				Sid:    "",
				Principal: struct {
					Federated string `json:"Federated,omitempty"`
				}{
					Federated: cfg.OrgsIAMOidcProviderArn,
				},
				Condition: struct {
					StringEquals map[string]string `json:"StringEquals"`
				}{
					StringEquals: map[string]string{
						fmt.Sprintf("%s:sub", cfg.OrgsIAMOidcProviderURL): fmt.Sprintf("system:serviceaccount:%s:%s", orgID, runnerOdrServiceAccountName(orgID)),
					},
				},
			},
		},
	}

	byts, err := json.Marshal(trustPolicy)
	if err != nil {
		return nil, fmt.Errorf("unable to create trust policy: %w", err)
	}

	return byts, nil
}
