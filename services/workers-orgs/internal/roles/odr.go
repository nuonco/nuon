package roles

import (
	"encoding/json"
	"fmt"
)

func OdrIAMName(orgID string) string {
	return fmt.Sprintf("org-odr-%s", orgID)
}

func runnerOdrServiceAccountName(orgID string) string {
	return fmt.Sprintf("waypoint-odr-%s", orgID)
}

// OdrIAMPolicy generates the policy for the deployment role
// bucketName is expected to be the deployments bucket in the orgs account
// orgID is expected to be the shortID of the org for this role
func OdrIAMPolicy(ecrRegistryARN, orgID string) ([]byte, error) {
	policy := iamRolePolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleStatement{
			// allow the role to read/write any of the orgs' repos
			{
				Effect: "Allow",
				Action: []string{
					"ecr:*",
				},
				Resource: fmt.Sprintf("%s/%s/*", ecrRegistryARN, orgID),
			},
			// allow the role to generate an token for any registry
			// this is relatively safe as it doesn't inherently give them permission for anything else
			{
				Effect: "Allow",
				Action: []string{
					"ecr:GetAuthorizationToken",
				},
				Resource: "*",
			},
			{
				Effect: "Allow",
				Action: []string{
					"sts:AssumeRole",
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

// OdrIAMTrustPolicy generates the trust policy for the ODR role
// The trust policy gives access to the runner service account of the cluster
// that the oidcProviderARN and oidcProviderURL belong to
func OdrIAMTrustPolicy(oidcProviderARN, oidcProviderURL, orgID string) ([]byte, error) {
	conditionKey := fmt.Sprintf("%s:sub", oidcProviderURL)
	conditionValue := fmt.Sprintf("system:serviceaccount:%s:%s", orgID, runnerOdrServiceAccountName(orgID))
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{"sts:AssumeRoleWithWebIdentity"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					Federated: oidcProviderARN,
				},
				Condition: iamCondition{
					StringEquals: map[string]string{
						conditionKey: conditionValue,
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
