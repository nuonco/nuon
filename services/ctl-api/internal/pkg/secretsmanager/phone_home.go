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

// ReconcilePhoneHomeKeyPolicy is not implemented yet.
//
// TODO(phone-home-auth): the shared CMK in the management account needs one key
// policy statement per distinct target account, granting kms:Decrypt and
// kms:DescribeKey to the account root, conditioned on
// kms:ViaService = secretsmanager.<management_region>.amazonaws.com — note that is
// the region of the *secret*, not the customer's install region, so one statement
// covers customers in every region.
//
// Three hazards to build for, none of which are addressed here:
//
//  1. PutKeyPolicy is a full replacement, not an append. Two installs provisioning
//     concurrently for different target accounts will read-modify-write the same
//     policy and one statement will be lost. Serialize the update (an activity keyed
//     on the CMK, or a Postgres advisory lock) and make it a reconcile-from-database
//     so a lost write self-heals on the next provision.
//  2. Reconcile on the union of target accounts, never per install. A statement for
//     account A must exist iff at least one non-deleted install targets A — pruning
//     when a single install is deleted would cut off sibling installs in the same
//     account, including any leftover deployed stack still being updated.
//  3. The key policy has a 32 KB ceiling, roughly low hundreds of accounts at ~200
//     bytes per statement. Emit a metric on policy size and alert well before it.
//     The escape hatch is sharding onto a second CMK keyed by
//     hash(target_account_id), which is why PhoneHomeAuth records the CMK ARN that
//     encrypted each secret.
//
// Blocked on the out-of-repo CMK and the ctl-api management-role IAM grants
// (rollout step 1). aws-sdk-go-v2/service/kms is not yet a dependency.
func ReconcilePhoneHomeKeyPolicy() error {
	return nil
}
