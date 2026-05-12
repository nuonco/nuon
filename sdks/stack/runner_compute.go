package stack

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sort"

	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	astypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// runnerInstanceType matches the TF default. Why: the init script's resource
// expectations (CPU/RAM headroom for the runner agent) are tuned for t3.medium.
const runnerInstanceType = "t3.medium"

// amazonOwnerID is Amazon's canonical owner ID for AL2023 AMIs.
const amazonOwnerID = "137112412989"

// runnerUserDataTemplate is byte-for-byte the TF user_data: wait for outbound
// HTTPS to come up (NAT route may lag instance boot) before sourcing the
// init script. The init script reads instance tags via ec2:DescribeTags.
const runnerUserDataTemplate = `#!/bin/bash
set -e
export RUNNER_AUTH_METHOD=iid
for i in $(seq 1 30); do
  if curl -fsS --max-time 5 -o /dev/null https://raw.githubusercontent.com; then
    break
  fi
  echo "waiting for outbound egress... ($i/30)"
  sleep 10
done
curl -fsSL https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init-mng-v2.sh | bash
`

// ensureRunnerCompute provisions the CloudWatch log group, looks up the AMI,
// then ensures the launch template + ASG exist. Idempotent: each step
// reconciles to the desired state.
//
// When refresh is true the runner instance is cycled at the end via
// StartInstanceRefresh. Why: a reprovision that touches no AWS resources
// still wants the running instance to pick up new tag values, refreshed AMI
// versions, or an updated init script — none of which a reboot replays.
// Provision passes refresh=true on first creation as a no-op (no instance
// exists to cycle); reprovision passes true to force a roll; deprovision
// shouldn't call this at all.
func ensureRunnerCompute(ctx context.Context, log *slog.Logger, ec2c *ec2.Client, iamc *iam.Client, asgc *autoscaling.Client, logsc *cloudwatchlogs.Client, st *State, cfg *Config, refresh bool) error {
	prefix := cfg.Prefix()

	// Block on IAM propagation before referencing the instance profile in the
	// launch template / ASG. Why: CreateInstanceProfile + AddRoleToProfile are
	// eventually consistent — for the first ~5–30s after the iam step
	// completes, EC2/ASG see the profile as "Invalid IAM Instance Profile
	// name" even though IAM itself reports it. Also catches the case where
	// state has a stale profile name from a different AWS account; in that
	// case the wait times out and the step fails with a clear error.
	if err := waitForInstanceProfile(ctx, log, iamc, st.RunnerInstanceProfileName, st.RunnerRoleName); err != nil {
		return err
	}

	if err := ensureLogGroup(ctx, log, logsc, st, prefix); err != nil {
		return err
	}
	amiID, err := lookupLatestAL2023AMI(ctx, ec2c)
	if err != nil {
		return err
	}
	if err := ensureLaunchTemplate(ctx, log, ec2c, st, cfg, amiID); err != nil {
		return err
	}
	if err := ensureASG(ctx, log, asgc, st, cfg); err != nil {
		return err
	}
	if refresh {
		if err := startInstanceRefresh(ctx, log, asgc, st.RunnerASGName); err != nil {
			return err
		}
	}
	return st.Save()
}

// retryOnInvalidInstanceProfile retries fn while AWS reports the IAM
// instance profile as invalid. ASG/EC2 see fresh profiles ~30–90s after
// IAM CreateInstanceProfile; this wraps the call so that lag doesn't fail
// the step. Caps at 3 minutes total wall time, 10s between attempts.
func retryOnInvalidInstanceProfile(ctx context.Context, log *slog.Logger, fn func() error) error {
	deadline := time.Now().Add(3 * time.Minute)
	logged := false
	for {
		err := fn()
		if err == nil {
			return nil
		}
		if !isInvalidInstanceProfileErr(err) {
			return err
		}
		if time.Now().After(deadline) {
			return err
		}
		if !logged {
			log.Info("waiting for IAM instance profile to propagate to EC2/ASG", "err", err.Error())
			logged = true
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}

// isInvalidInstanceProfileErr matches both the EC2-style and ASG-style
// validation errors that surface during IAM→EC2 propagation. Both come back
// as smithy.APIError with code "ValidationError" and a recognizable message;
// we substring-match on the stable phrase.
func isInvalidInstanceProfileErr(err error) bool {
	if err == nil {
		return false
	}
	if !isAWSErrCode(err, "ValidationError") {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Invalid IAM Instance Profile name") ||
		strings.Contains(msg, "Invalid IAM Instance Profile ARN")
}

// startInstanceRefresh cycles the running runner instance so it picks up
// the latest launch-template version (new tags, AMI, user-data). Idempotent
// in two senses: skips when an InstanceRefresh is already in progress, and
// no-ops when the ASG has no running instances yet (initial provision).
func startInstanceRefresh(ctx context.Context, log *slog.Logger, c *autoscaling.Client, asgName string) error {
	if asgName == "" {
		return nil
	}
	desc, err := c.DescribeInstanceRefreshes(ctx, &autoscaling.DescribeInstanceRefreshesInput{
		AutoScalingGroupName: &asgName,
	})
	if err != nil {
		return fmt.Errorf("describe instance refreshes: %w", err)
	}
	for _, r := range desc.InstanceRefreshes {
		switch r.Status {
		case astypes.InstanceRefreshStatusPending,
			astypes.InstanceRefreshStatusInProgress,
			astypes.InstanceRefreshStatusCancelling:
			log.Info("instance refresh already in progress, skipping", "refresh_id", aws.ToString(r.InstanceRefreshId))
			return nil
		}
	}
	out, err := c.StartInstanceRefresh(ctx, &autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: &asgName,
		Strategy:             astypes.RefreshStrategyRolling,
		Preferences: &astypes.RefreshPreferences{
			// Size=1 ASG: 0% min healthy means terminate the old before
			// launching the new. 180s warmup matches the init script's
			// outbound-egress retry window plus a buffer for runner
			// registration.
			MinHealthyPercentage: aws.Int32(0),
			InstanceWarmup:       aws.Int32(180),
		},
	})
	if err != nil {
		// First-time provision: no instances yet → AWS returns this and we
		// continue. The ASG itself will launch the first instance.
		if isAWSErrCode(err, "InstanceRefreshInProgressFault") {
			return nil
		}
		return fmt.Errorf("start instance refresh: %w", err)
	}
	log.Info("started instance refresh", "refresh_id", aws.ToString(out.InstanceRefreshId))
	return nil
}

// waitForInstanceProfile polls until the named profile exists in IAM and has
// the expected role attached. Times out after 60s with a clear error so a
// stale state file (profile name from a deleted/foreign account) fails fast
// rather than passing the bad name to ASG.
func waitForInstanceProfile(ctx context.Context, log *slog.Logger, c *iam.Client, profileName, roleName string) error {
	if profileName == "" {
		return fmt.Errorf("waitForInstanceProfile: profile name is empty (iam step did not populate state)")
	}
	deadline := time.Now().Add(60 * time.Second)
	logged := false
	for {
		out, err := c.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: &profileName})
		if err == nil {
			for _, r := range out.InstanceProfile.Roles {
				if aws.ToString(r.RoleName) == roleName {
					return nil
				}
			}
		} else if !isAWSErrCode(err, "NoSuchEntity") {
			return fmt.Errorf("get instance profile %s: %w", profileName, err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("instance profile %s not visible after 60s — IAM may still be propagating, or the cached name points to a deleted/foreign profile (try `stack-manager deprovision` or delete the local state file to recover)", profileName)
		}
		if !logged {
			log.Info("waiting for IAM instance profile to propagate", "name", profileName)
			logged = true
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func ensureLogGroup(ctx context.Context, log *slog.Logger, c *cloudwatchlogs.Client, st *State, prefix string) error {
	name := "/nuon/" + prefix + "/runner"
	out, err := c.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePrefix: &name})
	if err != nil {
		return fmt.Errorf("describe log groups: %w", err)
	}
	found := false
	for _, g := range out.LogGroups {
		if aws.ToString(g.LogGroupName) == name {
			found = true
			break
		}
	}
	if !found {
		if _, err := c.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: &name,
			Tags:         map[string]string{installIDTagKey: st.InstallID},
		}); err != nil && !isAWSErrCode(err, "ResourceAlreadyExistsException") {
			return fmt.Errorf("create log group: %w", err)
		}
		log.Info("created log group", "name", name)
	}
	if _, err := c.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    &name,
		RetentionInDays: aws.Int32(30),
	}); err != nil {
		return fmt.Errorf("set log group retention: %w", err)
	}
	st.RunnerLogGroupName = name
	_ = cwltypes.LogGroup{} // touch import for future tag reads
	return nil
}

func lookupLatestAL2023AMI(ctx context.Context, c *ec2.Client) (string, error) {
	out, err := c.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Owners: []string{amazonOwnerID},
		Filters: []ec2types.Filter{
			{Name: aws.String("name"), Values: []string{"al2023-ami-2023.*-x86_64"}},
			{Name: aws.String("virtualization-type"), Values: []string{"hvm"}},
			{Name: aws.String("state"), Values: []string{"available"}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("describe images: %w", err)
	}
	if len(out.Images) == 0 {
		return "", fmt.Errorf("no AL2023 AMI found")
	}
	sort.Slice(out.Images, func(i, j int) bool {
		return aws.ToString(out.Images[i].CreationDate) > aws.ToString(out.Images[j].CreationDate)
	})
	return aws.ToString(out.Images[0].ImageId), nil
}

// runnerTagPairs returns the canonical runner-instance tag set used on both
// launch-template TagSpecifications and ASG propagating tags. Why a single
// helper: the init script reads `nuon_runner_id` and `nuon_runner_api_url`
// from instance tags; drift between LT and ASG produces silent boot failures.
func runnerTagPairs(prefix string, cfg *Config) [][2]string {
	return [][2]string{
		{"Name", prefix + "-runner"},
		{"nuon_runner_id", cfg.RunnerID},
		{"nuon_runner_api_url", cfg.RunnerAPIURL},
		{"nuon_install_id", cfg.InstallID},
		{installIDTagKey, cfg.InstallID},
	}
}

func ensureLaunchTemplate(ctx context.Context, log *slog.Logger, c *ec2.Client, st *State, cfg *Config, amiID string) error {
	name := cfg.Prefix() + "-runner"
	userData := base64.StdEncoding.EncodeToString([]byte(runnerUserDataTemplate))

	tagPairs := runnerTagPairs(cfg.Prefix(), cfg)
	instanceTags := make([]ec2types.Tag, 0, len(tagPairs))
	for _, kv := range tagPairs {
		k, v := kv[0], kv[1]
		instanceTags = append(instanceTags, ec2types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	ltData := &ec2types.RequestLaunchTemplateData{
		ImageId:      &amiID,
		InstanceType: ec2types.InstanceType(runnerInstanceType),
		UserData:     &userData,
		IamInstanceProfile: &ec2types.LaunchTemplateIamInstanceProfileSpecificationRequest{
			Name: &st.RunnerInstanceProfileName,
		},
		MetadataOptions: &ec2types.LaunchTemplateInstanceMetadataOptionsRequest{
			HttpEndpoint: ec2types.LaunchTemplateInstanceMetadataEndpointStateEnabled,
			HttpTokens:   ec2types.LaunchTemplateHttpTokensStateRequired,
		},
		NetworkInterfaces: []ec2types.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{{
			AssociatePublicIpAddress: aws.Bool(false),
			DeviceIndex:              aws.Int32(0),
			Groups:                   []string{st.RunnerSecurityGroupID},
		}},
		BlockDeviceMappings: []ec2types.LaunchTemplateBlockDeviceMappingRequest{{
			DeviceName: aws.String("/dev/xvda"),
			Ebs: &ec2types.LaunchTemplateEbsBlockDeviceRequest{
				VolumeSize: aws.Int32(30),
				VolumeType: ec2types.VolumeTypeGp3,
			},
		}},
		TagSpecifications: []ec2types.LaunchTemplateTagSpecificationRequest{
			{ResourceType: ec2types.ResourceTypeInstance, Tags: instanceTags},
			{ResourceType: ec2types.ResourceTypeVolume, Tags: instanceTags},
		},
	}

	out, err := c.DescribeLaunchTemplates(ctx, &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateNames: []string{name},
	})
	if err != nil && !isAWSErrCode(err, "InvalidLaunchTemplateName.NotFoundException") {
		return fmt.Errorf("describe launch template: %w", err)
	}
	if err == nil && len(out.LaunchTemplates) > 0 {
		st.RunnerLaunchTemplateID = aws.ToString(out.LaunchTemplates[0].LaunchTemplateId)
		// Reconcile by adding a new version + setting it default. Why: AMI
		// ID and user_data both drift across runs (latest AL2023, bumped
		// init script). Without a fresh version, ASG instance refresh
		// would re-launch the old config.
		v, err := c.CreateLaunchTemplateVersion(ctx, &ec2.CreateLaunchTemplateVersionInput{
			LaunchTemplateId:   &st.RunnerLaunchTemplateID,
			LaunchTemplateData: ltData,
		})
		if err != nil {
			return fmt.Errorf("create launch template version: %w", err)
		}
		ver := fmt.Sprintf("%d", aws.ToInt64(v.LaunchTemplateVersion.VersionNumber))
		if _, err := c.ModifyLaunchTemplate(ctx, &ec2.ModifyLaunchTemplateInput{
			LaunchTemplateId: &st.RunnerLaunchTemplateID,
			DefaultVersion:   &ver,
		}); err != nil {
			return fmt.Errorf("modify launch template default: %w", err)
		}
		return nil
	}
	cr, err := c.CreateLaunchTemplate(ctx, &ec2.CreateLaunchTemplateInput{
		LaunchTemplateName: &name,
		LaunchTemplateData: ltData,
		TagSpecifications: []ec2types.TagSpecification{
			tagSpec(ec2types.ResourceTypeLaunchTemplate, name, st),
		},
	})
	if err != nil {
		return fmt.Errorf("create launch template: %w", err)
	}
	st.RunnerLaunchTemplateID = aws.ToString(cr.LaunchTemplate.LaunchTemplateId)
	log.Info("created launch template", "id", st.RunnerLaunchTemplateID)
	return nil
}

func ensureASG(ctx context.Context, log *slog.Logger, c *autoscaling.Client, st *State, cfg *Config) error {
	name := cfg.Prefix() + "-runner-asg"
	st.RunnerASGName = name

	asgTags := buildASGTags(name, cfg)
	ltSpec := &astypes.LaunchTemplateSpecification{
		LaunchTemplateId: &st.RunnerLaunchTemplateID,
		Version:          aws.String("$Latest"),
	}

	out, err := c.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	})
	if err != nil {
		return fmt.Errorf("describe asg: %w", err)
	}
	if len(out.AutoScalingGroups) == 0 {
		// IAM→EC2 eventual-consistency: a fresh instance profile is visible
		// to IAM (and our preflight wait) within seconds, but ASG validates
		// the profile via EC2's cache which can lag 30–90s. Retry on the
		// specific ValidationError to ride out the gap rather than failing
		// the whole step on a transient.
		if err := retryOnInvalidInstanceProfile(ctx, log, func() error {
			_, err := c.CreateAutoScalingGroup(ctx, &autoscaling.CreateAutoScalingGroupInput{
				AutoScalingGroupName: &name,
				MinSize:              aws.Int32(1),
				MaxSize:              aws.Int32(1),
				DesiredCapacity:      aws.Int32(1),
				VPCZoneIdentifier:    aws.String(st.RunnerSubnetID),
				LaunchTemplate:       ltSpec,
				Tags:                 asgTags,
			})
			return err
		}); err != nil {
			return fmt.Errorf("create asg: %w", err)
		}
		log.Info("created ASG", "name", name)
		return nil
	}
	// ASG exists — update LT pointer and trigger an instance refresh so the
	// running instance picks up the new launch-template default version.
	if _, err := c.UpdateAutoScalingGroup(ctx, &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &name,
		LaunchTemplate:       ltSpec,
		MinSize:              aws.Int32(1),
		MaxSize:              aws.Int32(1),
		DesiredCapacity:      aws.Int32(1),
		VPCZoneIdentifier:    aws.String(st.RunnerSubnetID),
	}); err != nil {
		return fmt.Errorf("update asg: %w", err)
	}
	if _, err := c.CreateOrUpdateTags(ctx, &autoscaling.CreateOrUpdateTagsInput{Tags: asgTags}); err != nil {
		return fmt.Errorf("update asg tags: %w", err)
	}
	if _, err := c.StartInstanceRefresh(ctx, &autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: &name,
	}); err != nil {
		// InstanceRefreshInProgress is fine — one is already running.
		if !isAWSErrCode(err, "InstanceRefreshInProgress") {
			log.Warn("start instance refresh (continuing)", "err", err)
		}
	}
	return nil
}

func buildASGTags(asgName string, cfg *Config) []astypes.Tag {
	pairs := runnerTagPairs(cfg.Prefix(), cfg)
	tags := make([]astypes.Tag, 0, len(pairs))
	for _, kv := range pairs {
		k, v := kv[0], kv[1]
		tags = append(tags, astypes.Tag{
			Key:               aws.String(k),
			Value:             aws.String(v),
			ResourceId:        aws.String(asgName),
			ResourceType:      aws.String("auto-scaling-group"),
			PropagateAtLaunch: aws.Bool(true),
		})
	}
	return tags
}

// deleteRunnerCompute tears down ASG, launch template, then log group. ASG
// must drain before launch template can be deleted; ForceDelete=true skips
// the wait for instance termination.
func deleteRunnerCompute(ctx context.Context, log *slog.Logger, ec2c *ec2.Client, asgc *autoscaling.Client, logsc *cloudwatchlogs.Client, st *State) error {
	if st.RunnerASGName != "" {
		if _, err := asgc.DeleteAutoScalingGroup(ctx, &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: &st.RunnerASGName,
			ForceDelete:          aws.Bool(true),
		}); err != nil {
			log.Warn("delete asg", "name", st.RunnerASGName, "err", err)
		}
		st.RunnerASGName = ""
	}
	if st.RunnerLaunchTemplateID != "" {
		if _, err := ec2c.DeleteLaunchTemplate(ctx, &ec2.DeleteLaunchTemplateInput{
			LaunchTemplateId: &st.RunnerLaunchTemplateID,
		}); err != nil {
			log.Warn("delete launch template", "id", st.RunnerLaunchTemplateID, "err", err)
		}
		st.RunnerLaunchTemplateID = ""
	}
	if st.RunnerLogGroupName != "" {
		if _, err := logsc.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: &st.RunnerLogGroupName,
		}); err != nil {
			log.Warn("delete log group", "name", st.RunnerLogGroupName, "err", err)
		}
		st.RunnerLogGroupName = ""
	}
	return st.Save()
}
