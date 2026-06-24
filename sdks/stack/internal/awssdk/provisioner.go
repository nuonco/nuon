package awssdk

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// Provisioner provisions an install stack using the AWS SDK directly. It
// implements core.Provisioner.
type Provisioner struct {
	region string
	awsCfg aws.Config

	ec2c  *ec2.Client
	iamc  *iam.Client
	stsc  *sts.Client
	asgc  *autoscaling.Client
	logsc *cloudwatchlogs.Client
	smc   *secretsmanager.Client

	// accountID is captured during validateAWS so buildOutputs can resolve
	// IAM role ARNs without a second sts call on the failure path.
	accountID string
}

var _ core.Provisioner = (*Provisioner)(nil)

// ReportsOwnRun is false: there is no phone-home in the AWS SDK path, so the
// stack package reports the run via the stack-run endpoints.
func (p *Provisioner) ReportsOwnRun() bool { return false }

// New loads the default AWS config for the region and constructs the service
// clients the provisioner needs.
func New(ctx context.Context, region string) (*Provisioner, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return &Provisioner{
		region: region,
		awsCfg: awsCfg,
		ec2c:   ec2.NewFromConfig(awsCfg),
		iamc:   iam.NewFromConfig(awsCfg),
		stsc:   sts.NewFromConfig(awsCfg),
		asgc:   autoscaling.NewFromConfig(awsCfg),
		logsc:  cloudwatchlogs.NewFromConfig(awsCfg),
		smc:    secretsmanager.NewFromConfig(awsCfg),
	}, nil
}

// Provision runs the full provisioning sequence. State is persisted to disk
// after each successful step so a partial failure can be cleaned up by
// Deprovision.
func (p *Provisioner) Provision(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config, kind core.Kind) (*core.Outputs, error) {
	if cfg.AWS == nil {
		return nil, fmt.Errorf("aws sdk provisioner: config missing aws block")
	}
	st, err := loadState(cfg.InstallID, p.region)
	if err != nil {
		return nil, err
	}
	// Propagate identity + cluster name into State so resource tagging
	// callsites can read them without holding cfg.
	st.OrgID = cfg.OrgID
	st.AppID = cfg.AppID
	st.ClusterName = cfg.AWS.ClusterName
	if st.ClusterName == "" {
		st.ClusterName = cfg.InstallID
	}

	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *State) error
	}{
		{"validate-aws", p.validateAWS},
		{"create-vpc", func(ctx context.Context, l *slog.Logger, s *State) error {
			return createVPC(ctx, l, p.ec2c, s)
		}},
		{"create-iam", func(ctx context.Context, l *slog.Logger, s *State) error {
			return createIAMRoles(ctx, l, sysLog, p.iamc, p.stsc, s, cfg)
		}},
		{"create-secrets", func(ctx context.Context, l *slog.Logger, s *State) error {
			return ensureSecrets(ctx, l, p.smc, s, cfg)
		}},
		{"create-runner-compute", func(ctx context.Context, l *slog.Logger, s *State) error {
			// Cycle the running instance on every run — initial provision is a
			// no-op (no instance exists yet); reprovision picks up refreshed
			// tags / AMI / user-data without manual intervention.
			refresh := kind == core.KindProvision || kind == core.KindReprovision
			return ensureRunnerCompute(ctx, l, p.ec2c, p.iamc, p.asgc, p.logsc, s, cfg, refresh)
		}},
	}

	for _, s := range steps {
		l := log.With("step", s.name)
		l.Info("step starting")
		if err := s.run(ctx, l, st); err != nil {
			l.Error("step failed", "err", err.Error())
			return nil, fmt.Errorf("step %s: %w", s.name, err)
		}
		l.Info("step completed")
	}

	log.Info("provision complete", "runner_asg", st.RunnerASGName)
	return p.buildOutputs(st, cfg), nil
}

// Deprovision tears down everything in the state file in reverse order.
func (p *Provisioner) Deprovision(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config) error {
	st, err := loadState(cfg.InstallID, p.region)
	if err != nil {
		return err
	}
	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *State) error
	}{
		{"delete-runner-compute", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteRunnerCompute(ctx, l, p.ec2c, p.asgc, p.logsc, s)
		}},
		{"delete-secrets", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteSecrets(ctx, l, p.smc, s, cfg)
		}},
		{"delete-iam", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteIAMRoles(ctx, l, p.iamc, s)
		}},
		{"delete-vpc", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteVPC(ctx, l, p.ec2c, s)
		}},
	}
	for _, s := range steps {
		l := log.With("step", s.name)
		l.Info("step starting")
		if err := s.run(ctx, l, st); err != nil {
			l.Error("step failed", "err", err.Error())
			return fmt.Errorf("step %s: %w", s.name, err)
		}
		l.Info("step completed")
	}
	if err := st.Delete(); err != nil {
		return fmt.Errorf("delete state file: %w", err)
	}
	log.Info("deprovision complete")
	return nil
}

// Status loads the persisted state and resolves it into Outputs. It makes a
// single STS call to resolve the account id so role ARNs are populated.
func (p *Provisioner) Status(ctx context.Context, cfg *core.Config) (*core.Outputs, error) {
	st, err := loadState(cfg.InstallID, p.region)
	if err != nil {
		return nil, err
	}
	if p.accountID == "" {
		if out, err := p.stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}); err == nil {
			p.accountID = aws.ToString(out.Account)
		}
	}
	return p.buildOutputs(st, cfg), nil
}

func (p *Provisioner) validateAWS(ctx context.Context, log *slog.Logger, _ *State) error {
	out, err := p.stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("sts get-caller-identity: %w", err)
	}
	p.accountID = aws.ToString(out.Account)
	log.Info("aws caller identity",
		"account", p.accountID,
		"arn", aws.ToString(out.Arn),
	)
	return nil
}

// buildOutputs resolves the persisted state into the method-agnostic Outputs
// the public stack package turns into a phone-home payload. Resolves role
// names to ARNs using the account id captured in validateAWS.
func (p *Provisioner) buildOutputs(st *State, cfg *core.Config) *core.Outputs {
	roleARN := func(name string) string {
		if name == "" || p.accountID == "" {
			return ""
		}
		return fmt.Sprintf("arn:aws:iam::%s:role/%s", p.accountID, name)
	}
	instanceProfileARN := func(name string) string {
		if name == "" || p.accountID == "" {
			return ""
		}
		return fmt.Sprintf("arn:aws:iam::%s:instance-profile/%s", p.accountID, name)
	}

	breakGlass := map[string]string{}
	for _, n := range st.BreakGlassRoleNames {
		breakGlass[n] = roleARN(n)
	}
	customRoles := map[string]string{}
	for _, n := range st.CustomRoleNames {
		customRoles[n] = roleARN(n)
	}

	installInputs := map[string]string{}
	if cfg != nil && len(cfg.InstallInputs) > 0 {
		installInputs = cfg.InstallInputs
	}

	secretARNs := map[string]string{}
	for k, v := range st.SecretARNs {
		secretARNs[k] = v
	}

	return &core.Outputs{
		Cloud:         core.CloudAWS,
		InstallInputs: installInputs,
		AWS: &core.AWSOutputs{
			AccountID: p.accountID,
			Region:    p.region,

			VPCID:                 st.VPCID,
			RunnerSubnetID:        st.RunnerSubnetID,
			PublicSubnetIDs:       st.PublicSubnetIDs,
			PrivateSubnetIDs:      st.PrivateSubnetIDs,
			RunnerSecurityGroupID: st.RunnerSecurityGroupID,

			RunnerIAMRoleARN:         roleARN(st.RunnerRoleName),
			RunnerInstanceProfileARN: instanceProfileARN(st.RunnerInstanceProfileName),
			RunnerASGName:            st.RunnerASGName,
			RunnerLogGroupName:       st.RunnerLogGroupName,

			ProvisionRoleARN:   roleARN(st.ProvisionRoleName),
			MaintenanceRoleARN: roleARN(st.MaintenanceRoleName),
			DeprovisionRoleARN: roleARN(st.DeprovisionRoleName),
			BreakGlassRoleARNs: breakGlass,
			CustomRoleARNs:     customRoles,

			SecretARNs: secretARNs,
		},
	}
}
