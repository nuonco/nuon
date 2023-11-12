package roles

import (
	"encoding/json"
	"fmt"
)

func InstallerIAMName(orgID string) string {
	return fmt.Sprintf("org-installer-%s", orgID)
}

// InstallerIAMPolicy generates the policy for actually performing installations. This is the role that the worker will
// assume before running both install.Provision or install.Deprovision.
func InstallerIAMPolicy(bucketName string, orgID string) ([]byte, error) {
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

			// allow the installer to assume an outside, third party IAM role
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

// TODO(jdt): is there a way we can restrict this to fewer services / roles?
// InstallerIAMTrustPolicy generates the trust policy for the installer role
// The trust policy gives access to any role arn with the provided prefix, in this case the EKS roles for our workers
// running in the main accounts.
func InstallerIAMTrustPolicy(workerRoleArnPrefix, supportRoleArn, odrRoleArn string) ([]byte, error) {
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
