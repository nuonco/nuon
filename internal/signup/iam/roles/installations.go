package roles

import (
	"encoding/json"
	"fmt"
)

func InstallationsIAMName(orgID string) string {
	return fmt.Sprintf("org-installations-access-%s", orgID)
}

// InstallationsIAMPolicy generates the policy for the installations role
// bucketName is expected to be the install bucket in the orgs account(?)
// orgID is expected to be the shortID of the org for this role
func InstallationsIAMPolicy(bucketName string, orgID string) ([]byte, error) {
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
// InstallationsIAMTrustPolicy generates the trust policy for the installations role
// The trust policy gives access to any service account in the default namespace of the cluster
// that the oidcProviderARN and oidcProviderURL belong to
func InstallationsIAMTrustPolicy(oidcProviderARN, oidcProviderURL string) ([]byte, error) {
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
