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
		Statement: []iamRoleStatement{
			// allow the role to read/write the orgID prefix of the bucketName bucket
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

// TODO(jdt): is there a way we can restrict this to fewer services / roles?
// DeploymentsIAMTrustPolicy generates the trust policy for the deployments role
// The trust policy gives access to any role arn with the provided prefix, in this case the EKS roles for our workers
// running in the main accounts.
func DeploymentsIAMTrustPolicy(workerRoleArnPrefix, supportRoleArn, odrRoleArn string) ([]byte, error) {
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: "*",
				},
				Condition: iamCondition{
					StringLike: map[string]string{
						"aws:PrincipalArn": workerRoleArnPrefix,
					},
				},
			},
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: "*",
				},
				Condition: iamCondition{
					StringEquals: map[string]string{
						"aws:PrincipalArn": odrRoleArn,
					},
				},
			},
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: "*",
				},
				Condition: iamCondition{
					StringEquals: map[string]string{
						"aws:PrincipalArn": supportRoleArn,
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
