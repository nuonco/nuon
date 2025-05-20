package roles

import (
	"encoding/json"
	"fmt"
)

func InstancesIAMName(orgID string) string {
	return fmt.Sprintf("org-instances-access-%s", orgID)
}

// InstancesIAMPolicy generates the policy for the instance role
// orgID is expected to be the shortID of the org for this role
func InstancesIAMPolicy(orgID string) ([]byte, error) {
	policy := iamRolePolicy{
		Version: defaultIAMPolicyVersion,

		Statement: []iamRoleStatement{
			// allow the role to read/write the ecr repositories for the org
			// predicated on tagging the repositories with the orgID
			{
				Effect: "Allow",
				Action: []string{
					"ecr:*",
				},
				Resource: "*",
				Condition: iamCondition{
					StringEquals: map[string]string{
						"ecr:ResourceTag/org-id": orgID,
					},
				},
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
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("unable to convert policy to json: %w", err)
	}
	return byts, nil
}

// This is used to write outputs to the instances bucket
func InstancesIAMTrustPolicy(runnerRoleARN, supportRoleArn, workersIAMRoleArnPrefix string) ([]byte, error) {
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: []string{"*"},
				},
				Condition: iamCondition{
					StringEquals: map[string]string{
						"aws:PrincipalArn": runnerRoleARN,
					},
				},
			},
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: []string{"*"},
				},
				Condition: iamCondition{
					StringEquals: map[string]string{
						"aws:PrincipalArn": supportRoleArn,
					},
				},
			},
			{
				Action: []string{"sts:AssumeRole"},
				Effect: "Allow",
				Sid:    "",
				Principal: iamPrincipal{
					AWS: []string{"*"},
				},
				Condition: iamCondition{
					StringLike: map[string]string{
						"aws:PrincipalArn": workersIAMRoleArnPrefix,
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
