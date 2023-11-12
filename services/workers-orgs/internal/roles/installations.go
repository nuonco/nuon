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
func InstallationsIAMPolicy(bucketName string, orgID string, sandboxBucketARN, sandboxKeyARN string) ([]byte, error) {
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
			// allow the role to read the sandbox bucket
			{
				Effect: "Allow",
				Action: []string{
					"s3:ListBucket",
				},
				Resource: sandboxBucketARN,
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetObject",
				},
				Resource: fmt.Sprintf("%s/sandboxes/*", sandboxBucketARN),
			},
			{
				Effect: "Allow",
				Action: []string{
					"kms:Decrypt",
				},
				Resource: sandboxKeyARN,
			},
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("unable to convert policy to json: %w", err)
	}
	return byts, nil
}

// InstallationsIAMTrustPolicy generates the trust policy for the installations role
// The trust policy gives access to any role arn with the provided prefix, in this case the EKS roles for our workers
// running in the main accounts.
func InstallationsIAMTrustPolicy(workerRoleArnPrefix, supportRoleArn, runnerRoleArn string) ([]byte, error) {
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
						"aws:PrincipalArn": runnerRoleArn,
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
