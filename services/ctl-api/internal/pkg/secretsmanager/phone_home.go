package secretsmanager

import (
	"encoding/json"
	"fmt"
)

// PhoneHomeSecretName is deterministic, unlike the ARN it resolves to.
func PhoneHomeSecretName(installID string) string {
	return fmt.Sprintf("nuon/phone-home/%s", installID)
}

// Tag keys on the phone-home secret. The secret name carries only the install ID, so
// without these there is no way to answer "which org owns this?" or "is this from
// staging?" from the AWS console or a cost report — every secret in the management
// account looks identical apart from an opaque ID.
//
// The domain-qualified form matches how Nuon labels resources elsewhere; runner_api_url
// and env are plain because they are environment facts rather than entity references.
const (
	TagKeyOrgID        = "org.nuon.co/id"
	TagKeyInstallID    = "install.nuon.co/id"
	TagKeyRunnerAPIURL = "runner_api_url"
	TagKeyEnv          = "env"
)

// PhoneHomeSecretTags identifies which install and org a secret belongs to, and which
// control plane created it.
//
// runner_api_url and env together disambiguate control planes that share a management
// account: a dev, staging and production ctl-api all write secrets named
// nuon/phone-home/<install_id>, and install IDs do not collide but nothing else
// distinguishes who owns a given entry. Empty values are dropped rather than written as
// empty tags.
func PhoneHomeSecretTags(orgID, installID, runnerAPIURL, env string) map[string]string {
	tags := map[string]string{}
	for k, v := range map[string]string{
		TagKeyOrgID:        orgID,
		TagKeyInstallID:    installID,
		TagKeyRunnerAPIURL: runnerAPIURL,
		TagKeyEnv:          env,
	} {
		if v != "" {
			tags[k] = v
		}
	}

	return tags
}

// PhoneHomeResourcePolicy grants the install's phone-home role read access to its
// own secret. Per-install isolation comes from this policy rather than from the CMK
// key policy, which is why one shared key is sufficient.
//
// The role is named in a Condition rather than as the Principal, because the role
// does not exist when this first runs and Secrets Manager validates principals:
// naming it directly fails the whole call with
//
//	MalformedPolicyDocumentException: This resource policy contains an unsupported principal
//
// The ordering that forces this cannot be reversed. The role is created by the
// customer's CloudFormation stack, but the Lambda reads this secret *during* that
// stack's creation, so the policy has to be in place first. (An install whose stack
// already ran does have the role, which is why this only ever failed on fresh
// installs — a difference worth knowing about before "it worked before" is taken as
// evidence of a regression.)
//
// The account root always resolves, so it is the Principal, and aws:PrincipalArn
// pins the grant to the one role at evaluation time. The effective boundary is the
// same as naming the role directly, with one caveat: root as Principal delegates to
// the target account's own IAM, so a principal there needs an identity policy
// allowing GetSecretValue on this ARN as well. The stack grants exactly that to the
// phone-home role and nothing else (getRunnerPhoneHomeLambdaRole), so reaching the
// secret still takes both halves — but a customer administrator in that account
// could write themselves a third. Account-level containment is the real boundary
// here, and the target account is the customer's own.
//
// A typo in the role name still fails silently, as an AccessDeniedException at
// phone-home time rather than an error here, which is why the render tests assert
// the name.
func PhoneHomeResourcePolicy(targetAccountID, phoneHomeRoleName string) (string, error) {
	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", targetAccountID, phoneHomeRoleName)

	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect": "Allow",
				"Principal": map[string]any{
					"AWS": fmt.Sprintf("arn:aws:iam::%s:root", targetAccountID),
				},
				"Action":   []string{"secretsmanager:GetSecretValue"},
				"Resource": "*",
				"Condition": map[string]any{
					"ArnEquals": map[string]any{
						"aws:PrincipalArn": roleARN,
					},
				},
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
