package roles

import (
	"encoding/json"
	"fmt"
)

func OrgsIAMName(orgID string) string {
	return fmt.Sprintf("org-orgs-access-%s", orgID)
}

// OrgsIAMPolicy generates the policy for accessing the orgid=%s namespace of the orgs bucket
func OrgsIAMPolicy(bucketName string, orgID string) ([]byte, error) {
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
// OrgsIAMTrustPolicy generates the trust policy for the orgs role, allowing it to be assumed by services in our
// cluster.
func OrgsIAMTrustPolicy(workerRoleArnPrefix, supportRoleArn, odrRoleArn string) ([]byte, error) {
	trustPolicy := iamRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []iamRoleTrustStatement{
			{
				Action: []string{"sts:AssumeRoleWithWebIdentity"},
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
				Action: []string{"sts:AssumeRoleWithWebIdentity"},
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
