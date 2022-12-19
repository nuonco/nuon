package roles

import (
	"encoding/json"
	"fmt"
)

const defaultIAMPolicyVersion string = "2012-10-17"

func DeploymentsIAMName(orgID string) string {
	return fmt.Sprintf("org-deployments-access-%s", orgID)
}

// DeploymentsIAMPolicy generates the policy for the deployment role
// bucketName is expected to be the deployments bucket in the orgs account
// orgID is expected to be the shortID of the org for this role
func DeploymentsIAMPolicy(bucketName string, orgID string) ([]byte, error) {
	policy := iamRolePolicy{
		Version: defaultIAMPolicyVersion,
		// allow the role to read/write the orgID prefix of the bucketName bucket
		Statement: []iamRoleStatement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:*",
				},
				Resource: fmt.Sprintf("arn:aws:s3:::%s/orgID=%s/*", bucketName, orgID),
			},
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("unable to convert policy to json: %w", err)
	}
	return byts, nil
}

// TODO(jdt): figure out how to restrict this to specific service accounts
// DeploymentsIAMTrustPolicy generates the trust policy for the deployments role
// The trust policy gives access to any service account in the default namespace of the cluster
// that the oidcProviderARN and oidcProviderURL belong to
func DeploymentsIAMTrustPolicy(oidcProviderARN, oidcProviderURL string) ([]byte, error) {
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
					Federated: oidcProviderARN,
				},
				Condition: struct {
					StringEquals map[string]string `json:"StringEquals"`
				}{
					StringEquals: map[string]string{
						fmt.Sprintf("%s:sub", oidcProviderURL): "system:serviceaccount:default:*",
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
