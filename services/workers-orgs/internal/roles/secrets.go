package roles

import (
	"encoding/json"
	"fmt"
)

// SecretsIAMName is the name of the policy / role
func SecretsIAMName(orgID string) string {
	return fmt.Sprintf("org-secrets-access-%s", orgID)
}

func SecretsKMSKeyPolicy(keyValuesRoleARN, currentServiceRoleARN, rootAccountARN string) ([]byte, error) {
	policy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{
					"kms:Create*",
					"kms:Describe*",
					"kms:Enable*",
					"kms:List*",
					"kms:Put*",
					"kms:Update*",
					"kms:Revoke*",
					"kms:Disable*",
					"kms:Get*",
					"kms:Delete*",
					"kms:ScheduleKeyDeletion",
					"kms:CancelKeyDeletion",
				},
				Effect:   "Allow",
				Resource: "*",
				Sid:      "Allow administration of the key",
				Principal: iamPrincipal{
					AWS: currentServiceRoleARN,
				},
			},
			{
				Action: []string{
					"kms:Create*",
					"kms:Describe*",
					"kms:Enable*",
					"kms:List*",
					"kms:Put*",
					"kms:Update*",
					"kms:Revoke*",
					"kms:Disable*",
					"kms:Get*",
					"kms:Delete*",
					"kms:ScheduleKeyDeletion",
					"kms:CancelKeyDeletion",
				},
				Effect:   "Allow",
				Resource: "*",
				Sid:      "Allow administration of the key",
				Principal: iamPrincipal{
					AWS: rootAccountARN,
				},
			},
			{
				Action:   []string{"kms:*"},
				Effect:   "Allow",
				Resource: "*",
				Sid:      "",
				Principal: iamPrincipal{
					AWS: "*",
				},
				Condition: iamCondition{
					StringLike: map[string]string{
						"aws:PrincipalArn": keyValuesRoleARN,
					},
				},
			},
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("unable to create kms key policy: %w", err)
	}

	return byts, nil
}

// SecretsIAMPolicy generates the policy for the key values role. It's worth noting, the secrets IAM policy is
// created before the key, and thus we do not know the arn of the key at the time of creation.
//
// However, the KMS key policy allows access to the the key-value IAM policy by arn, so practically it's not a huge
// problem.
func SecretsIAMPolicy(bucketName string, orgID string) ([]byte, error) {
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
			{
				Effect: "Allow",
				Action: []string{
					"kms:*",
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

// SecretsTrustPolicy is the trust policy that controls who can assume the key values IAM role
func SecretsTrustPolicy(workerRoleArnPrefix, supportRoleArn, odrRoleArn string) ([]byte, error) {
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
