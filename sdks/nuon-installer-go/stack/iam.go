package stack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Trust policy snippets that mirror install-stacks/aws/iam.tf.

const ec2AssumePolicy = `{
  "Version": "2012-10-17",
  "Statement": [{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]
}`

// CreateIAMRoles is the SDK equivalent of install-stacks/aws/iam.tf. It
// provisions, in order:
//  1. The runner role + instance profile (EC2 trust).
//  2. Operation roles (provision/maintenance/deprovision), each conditional
//     on having an inline policy or managed-policy ARNs configured.
//  3. Dynamic break-glass and custom roles, keyed verbatim by map key.
//  4. The runner inline policy (built last because its AssumeRole resources
//     reference whichever ops/break-glass/custom roles ended up existing).
func CreateIAMRoles(ctx context.Context, log *slog.Logger, c *iam.Client, stsc *sts.Client, st *State, cfg *Config) error {
	accountID, err := callerAccountID(ctx, stsc)
	if err != nil {
		return err
	}
	prefix := cfg.Prefix()

	// Reset role-name fields each run; we'll repopulate based on what cfg
	// requested. Why: a config that previously enabled provision but no
	// longer does should fall through to the "no role" branch in phone-home.
	st.ProvisionRoleName = ""
	st.MaintenanceRoleName = ""
	st.DeprovisionRoleName = ""
	st.BreakGlassRoleNames = nil
	st.CustomRoleNames = nil

	// 1. Runner role + instance profile.
	st.RunnerRoleName = prefix + "-runner"
	if err := ensureRole(ctx, log, c, st.RunnerRoleName, ec2AssumePolicy, st.InstallID); err != nil {
		return err
	}
	st.RunnerInstanceProfileName = prefix + "-runner"
	if err := ensureInstanceProfile(ctx, log, c, st.RunnerInstanceProfileName, st.RunnerRoleName, st.InstallID); err != nil {
		return err
	}

	// 2. Operation roles. The control_plane_assume trust references the
	// runner role's ARN, so build it after the runner role exists.
	runnerARN := iamRoleARN(accountID, st.RunnerRoleName)
	trust := buildControlPlaneAssume(accountID, runnerARN, cfg.NuonSupportIAMRoleARNs)

	provisionInline := resolveInline(cfg.ProvisionInlinePolicyDocument, cfg.ProvisionPermissions)
	if provisionInline != "" || len(cfg.ProvisionManagedPolicyARNs) > 0 {
		st.ProvisionRoleName = prefix + "-provision"
		if err := ensureRoleWithPolicies(ctx, log, c, st.ProvisionRoleName, trust, provisionInline,
			prefix+"-provision-inline", cfg.ProvisionManagedPolicyARNs, st.InstallID); err != nil {
			return err
		}
	}
	maintenanceInline := resolveInline(cfg.MaintenanceInlinePolicyDocument, cfg.MaintenancePermissions)
	if maintenanceInline != "" || len(cfg.MaintenanceManagedPolicyARNs) > 0 {
		st.MaintenanceRoleName = prefix + "-maintenance"
		if err := ensureRoleWithPolicies(ctx, log, c, st.MaintenanceRoleName, trust, maintenanceInline,
			prefix+"-maintenance-inline", cfg.MaintenanceManagedPolicyARNs, st.InstallID); err != nil {
			return err
		}
	}
	deprovisionInline := resolveInline(cfg.DeprovisionInlinePolicyDocument, cfg.DeprovisionPermissions)
	if deprovisionInline != "" || len(cfg.DeprovisionManagedPolicyARNs) > 0 {
		st.DeprovisionRoleName = prefix + "-deprovision"
		if err := ensureRoleWithPolicies(ctx, log, c, st.DeprovisionRoleName, trust, deprovisionInline,
			prefix+"-deprovision-inline", cfg.DeprovisionManagedPolicyARNs, st.InstallID); err != nil {
			return err
		}
	}

	// 3. Dynamic break-glass + custom roles. Keys come from cfg.* maps; role
	// names are taken verbatim — TF uses each.key without any wrapping to
	// stay under IAM's 64-char role-name limit, and we match.
	for _, k := range sortedEnabledKeys(cfg.BreakGlassRoles) {
		v := cfg.BreakGlassRoles[k]
		inline := resolveInline(v.InlinePolicyDocument, v.Permissions)
		if err := ensureRoleWithPolicies(ctx, log, c, k, trust, inline, k+"-inline", v.ManagedPolicyARNs, st.InstallID); err != nil {
			return err
		}
		st.BreakGlassRoleNames = append(st.BreakGlassRoleNames, k)
	}
	for _, k := range sortedEnabledKeys(cfg.CustomRoles) {
		v := cfg.CustomRoles[k]
		inline := resolveInline(v.InlinePolicyDocument, v.Permissions)
		if err := ensureRoleWithPolicies(ctx, log, c, k, trust, inline, k+"-inline", v.ManagedPolicyARNs, st.InstallID); err != nil {
			return err
		}
		st.CustomRoleNames = append(st.CustomRoleNames, k)
	}

	// 4. Runner inline policy. Built last so AssumeOperationRoles resources
	// reference roles that actually exist.
	assumeResources := []string{}
	for _, n := range []string{st.ProvisionRoleName, st.MaintenanceRoleName, st.DeprovisionRoleName} {
		if n != "" {
			assumeResources = append(assumeResources, iamRoleARN(accountID, n))
		}
	}
	for _, n := range st.BreakGlassRoleNames {
		assumeResources = append(assumeResources, iamRoleARN(accountID, n))
	}
	for _, n := range st.CustomRoleNames {
		assumeResources = append(assumeResources, iamRoleARN(accountID, n))
	}
	doc := buildRunnerInlineDoc(prefix, assumeResources)
	if err := putRolePolicy(ctx, c, st.RunnerRoleName, prefix+"-runner-inline", doc); err != nil {
		return err
	}

	return st.Save()
}

// ensureRole creates an IAM role with the given trust policy if it doesn't
// already exist. Idempotent.
func ensureRole(ctx context.Context, log *slog.Logger, c *iam.Client, name, trust, installID string) error {
	_, err := c.GetRole(ctx, &iam.GetRoleInput{RoleName: &name})
	if err == nil {
		return nil
	}
	if !IsAWSErrCode(err, "NoSuchEntity") {
		return fmt.Errorf("get role %s: %w", name, err)
	}
	if _, err := c.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 &name,
		AssumeRolePolicyDocument: &trust,
		Tags: []iamtypes.Tag{
			{Key: aws.String(installIDTagKey), Value: &installID},
		},
	}); err != nil {
		if !IsAWSErrCode(err, "EntityAlreadyExists") {
			return fmt.Errorf("create role %s: %w", name, err)
		}
		return nil
	}
	log.Info("created IAM role", "role", name)
	return nil
}

// ensureRoleWithPolicies provisions an IAM role then attaches/refreshes its
// inline policy and managed-policy attachments. Inline policy is omitted
// when inlineDoc is empty (matching the TF count=0 branch).
func ensureRoleWithPolicies(ctx context.Context, log *slog.Logger, c *iam.Client,
	name, trust, inlineDoc, inlineName string, managedARNs []string, installID string) error {
	if err := ensureRole(ctx, log, c, name, trust, installID); err != nil {
		return err
	}
	if inlineDoc != "" {
		if err := putRolePolicy(ctx, c, name, inlineName, inlineDoc); err != nil {
			return err
		}
	}
	return reconcileAttachedPolicies(ctx, c, name, managedARNs)
}

func ensureInstanceProfile(ctx context.Context, log *slog.Logger, c *iam.Client, name, role, installID string) error {
	_, err := c.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: &name})
	if err != nil {
		if !IsAWSErrCode(err, "NoSuchEntity") {
			return fmt.Errorf("get instance profile %s: %w", name, err)
		}
		if _, err := c.CreateInstanceProfile(ctx, &iam.CreateInstanceProfileInput{
			InstanceProfileName: &name,
			Tags: []iamtypes.Tag{
				{Key: aws.String(installIDTagKey), Value: &installID},
			},
		}); err != nil && !IsAWSErrCode(err, "EntityAlreadyExists") {
			return fmt.Errorf("create instance profile %s: %w", name, err)
		}
		log.Info("created instance profile", "name", name)
	}
	// Add role to profile if not already attached.
	out, err := c.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: &name})
	if err != nil {
		return fmt.Errorf("get instance profile after create: %w", err)
	}
	for _, r := range out.InstanceProfile.Roles {
		if aws.ToString(r.RoleName) == role {
			return nil
		}
	}
	if _, err := c.AddRoleToInstanceProfile(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: &name,
		RoleName:            &role,
	}); err != nil && !IsAWSErrCode(err, "LimitExceeded") {
		return fmt.Errorf("add role to profile: %w", err)
	}
	return nil
}

func putRolePolicy(ctx context.Context, c *iam.Client, role, name, doc string) error {
	_, err := c.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       &role,
		PolicyName:     &name,
		PolicyDocument: &doc,
	})
	if err != nil {
		return fmt.Errorf("put role policy %s on %s: %w", name, role, err)
	}
	return nil
}

// reconcileAttachedPolicies attaches any policies missing from the role.
// AttachRolePolicy is AWS-side idempotent, but skipping unnecessary calls
// keeps logs and CloudTrail tidy.
func reconcileAttachedPolicies(ctx context.Context, c *iam.Client, role string, want []string) error {
	if len(want) == 0 {
		return nil
	}
	out, err := c.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: &role})
	if err != nil {
		return fmt.Errorf("list attached policies for %s: %w", role, err)
	}
	have := make(map[string]bool, len(out.AttachedPolicies))
	for _, p := range out.AttachedPolicies {
		have[aws.ToString(p.PolicyArn)] = true
	}
	for _, p := range want {
		if have[p] {
			continue
		}
		p := p
		if _, err := c.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{RoleName: &role, PolicyArn: &p}); err != nil {
			return fmt.Errorf("attach %s -> %s: %w", p, role, err)
		}
	}
	return nil
}

// resolveInline mirrors install-stacks/aws/locals.tf: a non-empty inline
// document wins; otherwise build an Allow-on-`*` policy from permissions.
// Empty input returns "" (signaling "no inline policy on this role").
func resolveInline(doc string, perms []string) string {
	if doc != "" {
		return doc
	}
	if len(perms) == 0 {
		return ""
	}
	b, _ := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{{
			"Effect":   "Allow",
			"Action":   perms,
			"Resource": "*",
		}},
	})
	return string(b)
}

// buildControlPlaneAssume mirrors install-stacks/aws/iam.tf control_plane_assume.
// Statement 1 trusts Nuon support role ARNs (or falls back to the customer's
// account root); statement 2 trusts the runner role for in-account assumes.
func buildControlPlaneAssume(accountID, runnerARN string, supportARNs []string) string {
	principals := supportARNs
	if len(principals) == 0 {
		principals = []string{fmt.Sprintf("arn:aws:iam::%s:root", accountID)}
	}
	doc := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": map[string]any{"AWS": principals},
				"Action":    "sts:AssumeRole",
			},
			{
				"Effect":    "Allow",
				"Principal": map[string]any{"AWS": runnerARN},
				"Action":    "sts:AssumeRole",
			},
		},
	}
	b, _ := json.Marshal(doc)
	return string(b)
}

// buildRunnerInlineDoc mirrors data.aws_iam_policy_document.runner_inline.
// Why: the runner has to assume operation roles, read its own secrets,
// write its own logs, and DescribeTags so init-mng-v2.sh can resolve its
// runner_id / runner_api_url from instance tags.
func buildRunnerInlineDoc(prefix string, assumeResources []string) string {
	stmts := []map[string]any{}
	if len(assumeResources) > 0 {
		stmts = append(stmts, map[string]any{
			"Sid":      "AssumeOperationRoles",
			"Effect":   "Allow",
			"Action":   "sts:AssumeRole",
			"Resource": assumeResources,
		})
	}
	stmts = append(stmts,
		map[string]any{
			"Sid":      "ReadOwnSecrets",
			"Effect":   "Allow",
			"Action":   []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"},
			"Resource": fmt.Sprintf("arn:aws:secretsmanager:*:*:secret:%s-*", prefix),
		},
		map[string]any{
			"Sid":    "RunnerCloudWatchLogs",
			"Effect": "Allow",
			"Action": []string{
				"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents",
				"logs:DescribeLogStreams",
			},
			"Resource": []string{
				fmt.Sprintf("arn:aws:logs:*:*:log-group:/nuon/%s/*", prefix),
				"arn:aws:logs:*:*:log-group:runner-*",
				"arn:aws:logs:*:*:log-group:runner-*:*",
			},
		},
		map[string]any{
			"Sid":      "RunnerDescribeTags",
			"Effect":   "Allow",
			"Action":   "ec2:DescribeTags",
			"Resource": "*",
		},
	)
	b, _ := json.Marshal(map[string]any{
		"Version":   "2012-10-17",
		"Statement": stmts,
	})
	return string(b)
}

func iamRoleARN(accountID, name string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, name)
}

func callerAccountID(ctx context.Context, stsc *sts.Client) (string, error) {
	out, err := stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("sts get-caller-identity: %w", err)
	}
	return aws.ToString(out.Account), nil
}

// sortedEnabledKeys returns the enabled keys of a role map in stable
// (sorted) order so retries hit AWS in the same sequence and logs read
// deterministically across runs.
func sortedEnabledKeys(m map[string]RoleConfig) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if !v.Enabled {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// DeleteIAMRoles tears down everything CreateIAMRoles built. Idempotent:
// missing-role errors are swallowed so a partially-applied state still drains.
func DeleteIAMRoles(ctx context.Context, log *slog.Logger, c *iam.Client, st *State) error {
	// Detach the runner role from its instance profile + delete the profile.
	if st.RunnerInstanceProfileName != "" {
		if st.RunnerRoleName != "" {
			_, _ = c.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: &st.RunnerInstanceProfileName,
				RoleName:            &st.RunnerRoleName,
			})
		}
		if _, err := c.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: &st.RunnerInstanceProfileName}); err != nil && !IsAWSErrCode(err, "NoSuchEntity") {
			log.Warn("delete instance profile", "name", st.RunnerInstanceProfileName, "err", err)
		}
		st.RunnerInstanceProfileName = ""
	}

	roles := []string{
		st.RunnerRoleName,
		st.ProvisionRoleName,
		st.MaintenanceRoleName,
		st.DeprovisionRoleName,
	}
	roles = append(roles, st.BreakGlassRoleNames...)
	roles = append(roles, st.CustomRoleNames...)
	for _, name := range roles {
		if name == "" {
			continue
		}
		deleteRole(ctx, log, c, name)
	}
	st.RunnerRoleName = ""
	st.ProvisionRoleName = ""
	st.MaintenanceRoleName = ""
	st.DeprovisionRoleName = ""
	st.BreakGlassRoleNames = nil
	st.CustomRoleNames = nil
	return st.Save()
}

func deleteRole(ctx context.Context, log *slog.Logger, c *iam.Client, name string) {
	// Detach managed policies + delete inline policies first; IAM rejects
	// DeleteRole on a role with attachments.
	if out, err := c.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: &name}); err == nil {
		for _, p := range out.AttachedPolicies {
			_, _ = c.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{RoleName: &name, PolicyArn: p.PolicyArn})
		}
	}
	if out, err := c.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: &name}); err == nil {
		for _, p := range out.PolicyNames {
			p := p
			_, _ = c.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{RoleName: &name, PolicyName: &p})
		}
	}
	if _, err := c.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: &name}); err != nil && !IsAWSErrCode(err, "NoSuchEntity") {
		log.Warn("delete role", "role", name, "err", err)
	}
}
