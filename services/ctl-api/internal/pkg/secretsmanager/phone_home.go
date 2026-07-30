package secretsmanager

import (
	"encoding/json"
	"fmt"
)

// PhoneHomeSecretName is deterministic, unlike the ARN it resolves to.
func PhoneHomeSecretName(installID string) string {
	return fmt.Sprintf("nuon/phone-home/%s", installID)
}

// PhoneHomeResourcePolicy grants the install's phone-home role read access to its
// own secret. Per-install isolation comes from this policy rather than from the CMK
// key policy, which is why one shared key is sufficient.
//
// The named role does not exist yet when this is first applied — it is created by
// the CloudFormation stack the customer has not run. That is fine: an IAM resource
// policy may reference a principal ARN that does not exist, and simply matches
// nothing until it appears. The consequence is that a typo in the role name fails
// silently as an AccessDeniedException at phone-home time, which is why the render
// tests assert the name.
func PhoneHomeResourcePolicy(targetAccountID, phoneHomeRoleName string) (string, error) {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect": "Allow",
				"Principal": map[string]any{
					"AWS": fmt.Sprintf("arn:aws:iam::%s:role/%s", targetAccountID, phoneHomeRoleName),
				},
				"Action":   []string{"secretsmanager:GetSecretValue"},
				"Resource": "*",
			},
		},
	}

	byts, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("unable to marshal resource policy: %w", err)
	}

	return string(byts), nil
}

// Deliberately absent: anything that edits the CMK key policy.
//
// An earlier revision planned one key-policy statement per target account, added here
// as installs were provisioned. That cannot coexist with Terraform — PutKeyPolicy is a
// full replacement, so every apply would revert the grants and every reconcile would
// show as drift — and it caps out at the 32KB key-policy limit somewhere in the low
// hundreds of accounts.
//
// The key policy is now static and owned by Terraform
// (services/ctl-api/infra/phone_home.tf): one statement allows kms:Decrypt through
// Secrets Manager to any principal whose role is named like a phone-home Lambda role.
// The effective boundary is unchanged, because reaching the key at all requires
// GetSecretValue on a specific secret, which PhoneHomeResourcePolicy above grants to
// exactly one role. Do not reintroduce a KMS write path here.
