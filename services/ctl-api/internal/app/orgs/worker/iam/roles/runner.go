package roles

import (
	"encoding/json"
	"fmt"
)

func RunnerIAMName(orgID string) string {
	return fmt.Sprintf("runner-%s", orgID)
}

// RunnerIAMPolicy generates the policy for the deployment role
// bucketName is expected to be the deployments bucket in the orgs account
// orgID is expected to be the shortID of the org for this role
func RunnerIAMPolicy(ecrRegistryARN, orgID string) ([]byte, error) {
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
					"sts:AssumeRoleWithWebIdentity",
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

// RunnerIAMTrustPolicy generates the trust policy for the ODR role
// The trust policy gives access to the runner service account of the cluster
// that the oidcProviderARN and oidcProviderURL belong to
func RunnerIAMTrustPolicy(supportRoleARN, oidcProviderARN, oidcProviderURL, orgID string) ([]byte, error) {
	// NOTE(jm): this needs to be an OIDC provider url in the orgs account, that
	// delegates to the OIDC provider for the runner
	conditionKey := fmt.Sprintf("%s:sub", oidcProviderURL)

	// service accounts can be in any namespace in the orgs account, as they are per runner-group
	conditionValue := "system:serviceaccount:*:runner-*"

	// build out trust policy
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{"sts:AssumeRoleWithWebIdentity"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					// NOTE(jm): this needs to be an OIDC provider ARN in the same account, that
					// delegates to the OIDC provider for the runner
					Federated: oidcProviderARN,
				},
				Condition: iamCondition{
					StringLike: map[string]string{
						conditionKey: conditionValue,
					},
				},
			},
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: []string{supportRoleARN},
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
