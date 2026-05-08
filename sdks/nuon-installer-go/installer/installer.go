// Package installer orchestrates AWS-native install stack provisioning.
//
// It wires the AWS SDK clients to the stack package's resource provisioners
// and emits structured logs through the logstream package, which targets
// either stdout (dev) or the Nuon ctl-api OTLP log-stream ingest (production).
package installer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/sdks/nuon-installer-go/logstream"
	"github.com/nuonco/nuon/sdks/nuon-installer-go/stack"
	"github.com/nuonco/nuon/sdks/nuon-installer-go/stackrun"
)

// Installer provisions and tears down an install stack in a customer AWS account.
type Installer struct {
	opts Options
	log  *slog.Logger
	prov *logstream.Provider

	awsCfg aws.Config
	ec2c   *ec2.Client
	iamc   *iam.Client
	stsc   *sts.Client
	asgc   *autoscaling.Client
	logsc  *cloudwatchlogs.Client
	smc    *secretsmanager.Client

	// cfg is hydrated from the CreateRun response and threaded into every
	// stack-package call. nil until Provision/Deprovision fetches it.
	cfg *stack.Config

	// accountID is captured during stepValidateAWS so reportRun can build
	// IAM role ARNs without a second sts call on the failure path.
	accountID string
}

// New builds an Installer. Caller must call Close to flush logs.
func New(ctx context.Context, opts Options) (*Installer, error) {
	if opts.InstallID == "" {
		return nil, fmt.Errorf("InstallID required")
	}
	if opts.AWSRegion == "" {
		return nil, fmt.Errorf("AWSRegion required")
	}

	var prov *logstream.Provider
	if opts.LogStream == nil {
		prov = logstream.NewStdout("stack")
	} else {
		var err error
		prov, err = logstream.New(ctx, logstream.Config{
			RunnerAPIURL: opts.LogStream.RunnerAPIURL,
			LogStreamID:  opts.LogStream.ID,
			WriteToken:   opts.LogStream.WriteToken,
			ServiceName:  "stack",
			Attrs: map[string]string{
				"install_id": opts.InstallID,
				"aws_region": opts.AWSRegion,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(opts.AWSRegion))
	if err != nil {
		_ = prov.Shutdown(ctx)
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &Installer{
		opts:   opts,
		log:    prov.Logger().With("install_id", opts.InstallID, "aws_region", opts.AWSRegion),
		prov:   prov,
		awsCfg: awsCfg,
		ec2c:   ec2.NewFromConfig(awsCfg),
		iamc:   iam.NewFromConfig(awsCfg),
		stsc:   sts.NewFromConfig(awsCfg),
		asgc:   autoscaling.NewFromConfig(awsCfg),
		logsc:  cloudwatchlogs.NewFromConfig(awsCfg),
		smc:    secretsmanager.NewFromConfig(awsCfg),
	}, nil
}

func (i *Installer) Close(ctx context.Context) error {
	if i.prov != nil {
		return i.prov.Shutdown(ctx)
	}
	return nil
}

// Provision runs the full provisioning sequence. State is persisted to disk
// after each successful step so a partial failure can be cleaned up by Deprovision.
// Equivalent to Run(ctx, RunKindProvision).
func (i *Installer) Provision(ctx context.Context) (*stack.State, error) {
	return i.Run(ctx, stackrun.RunKindProvision)
}

// Reprovision is a re-run on an existing install. Functionally identical to
// Provision (both code paths are idempotent — every step is discover-or-create
// keyed on the install_id tag) but recorded as a distinct run kind so the
// dashboard can show first-time vs reconcile in the audit trail.
func (i *Installer) Reprovision(ctx context.Context) (*stack.State, error) {
	return i.Run(ctx, stackrun.RunKindReprovision)
}

// Run executes a provision-shaped workflow under the given kind. Exposed for
// callers that want to label the run explicitly; most should use the
// Provision / Reprovision wrappers.
func (i *Installer) Run(ctx context.Context, kind stackrun.RunKind) (*stack.State, error) {
	st, err := stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return nil, err
	}

	// Report to ctl-api as a stack run. Required when configured: the response
	// also carries the OTLP log-stream credentials and the rendered Config we
	// need for visibility and resource provisioning.
	var runClient *stackrun.Client
	var runID string
	if i.opts.StackRun != nil {
		runClient = stackrun.New(stackrun.Config{
			CtlAPIURL:   i.opts.StackRun.CtlAPIURL,
			InstallID:   i.opts.InstallID,
			PhoneHomeID: i.opts.StackRun.PhoneHomeID,
		})
		resp, err := runClient.CreateRun(ctx, kind)
		if err != nil {
			return st, fmt.Errorf("create stack run: %w", err)
		}
		runID = resp.ID
		i.log.Info("created stack run", "run_id", runID)
		if resp.Config == nil {
			return st, fmt.Errorf("create stack run: ctl-api returned no config block")
		}
		i.cfg = resp.Config
		i.cfg.InstallID = i.opts.InstallID
		// Validate the bootstrap fields before any AWS resources change. The
		// runner's init script reads nuon_runner_id / nuon_runner_api_url
		// from EC2 instance tags via IMDSv2 — empty values would let
		// provisioning succeed but leave the runner unable to authenticate,
		// surfacing as a silent "never connects" later.
		if i.cfg.RunnerID == "" {
			return st, fmt.Errorf("ctl-api config missing runner_id — install has no runner attached")
		}
		if i.cfg.RunnerAPIURL == "" {
			return st, fmt.Errorf("ctl-api config missing runner_api_url — set RunnerGroupSettings.RunnerAPIURL on this install's runner group")
		}
		// Swap the stdout-only provider for an OTLP one fed by the credentials
		// ctl-api just minted for this run.
		if resp.LogStream != nil && resp.LogStream.ID != "" {
			otelProv, err := logstream.New(ctx, logstream.Config{
				RunnerAPIURL: resp.LogStream.RunnerAPIURL,
				LogStreamID:  resp.LogStream.ID,
				WriteToken:   resp.LogStream.WriteToken,
				ServiceName:  "stack",
				Attrs: map[string]string{
					"install_id": i.opts.InstallID,
					"aws_region": i.opts.AWSRegion,
				},
			})
			if err != nil {
				i.log.Warn("init otlp log stream failed (continuing with prior logger)", "err", err.Error())
			} else {
				if i.prov != nil {
					_ = i.prov.Shutdown(ctx)
				}
				i.prov = otelProv
				i.log = otelProv.Logger().With(
					"install_id", i.opts.InstallID,
					"aws_region", i.opts.AWSRegion,
				)
			}
		}
	} else {
		// No StackRun configured (e.g. --stdout-only): run with an empty
		// config so VPC/log-group/secrets can still be exercised end-to-end.
		i.cfg = &stack.Config{InstallID: i.opts.InstallID}
	}

	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *stack.State) error
	}{
		{"validate-aws", i.stepValidateAWS},
		{"create-vpc", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateVPC(ctx, l, i.ec2c, s)
		}},
		{"create-iam", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateIAMRoles(ctx, l, i.iamc, i.stsc, s, i.cfg)
		}},
		{"create-secrets", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.EnsureSecrets(ctx, l, i.smc, s, i.cfg)
		}},
		{"create-runner-compute", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			// Cycle the running instance on every Run — initial provision is a
			// no-op (no instance exists yet); reprovision picks up refreshed
			// tags / AMI / user-data without manual intervention.
			refresh := kind == stackrun.RunKindProvision || kind == stackrun.RunKindReprovision
			return stack.EnsureRunnerCompute(ctx, l, i.ec2c, i.iamc, i.asgc, i.logsc, s, i.cfg, refresh)
		}},
	}

	for _, s := range steps {
		log := i.log.With("step", s.name)
		log.Info("step starting")
		if err := s.run(ctx, log, st); err != nil {
			log.Error("step failed", "err", err.Error())
			i.reportRun(ctx, runClient, runID, "failed", err.Error(), st)
			return st, fmt.Errorf("step %s: %w", s.name, err)
		}
		log.Info("step completed")
	}

	i.log.Info("provision complete", "runner_asg", st.RunnerASGName)
	i.reportRun(ctx, runClient, runID, "succeeded", "", st)
	return st, nil
}

// reportRun is best-effort; failures only log. Builds the phone-home payload
// described in install-stacks/aws/phone_home.tf so app templates resolving
// `nuon.install_stack.outputs.*` see identical key sets across CFN/TF/SDK.
func (i *Installer) reportRun(ctx context.Context, c *stackrun.Client, runID, status, statusDesc string, st *stack.State) {
	if c == nil || runID == "" {
		return
	}
	data := i.buildPhoneHomePayload(st)
	if err := c.UpdateRun(ctx, runID, stackrun.UpdateRunRequest{
		Status:            status,
		StatusDescription: statusDesc,
		Data:              data,
	}); err != nil {
		i.log.Warn("update stack run failed", "err", err.Error(), "status", status)
	}
}

func (i *Installer) buildPhoneHomePayload(st *stack.State) map[string]any {
	roleARN := func(name string) string {
		if name == "" || i.accountID == "" {
			return ""
		}
		return fmt.Sprintf("arn:aws:iam::%s:role/%s", i.accountID, name)
	}
	instanceProfileARN := func(name string) string {
		if name == "" || i.accountID == "" {
			return ""
		}
		return fmt.Sprintf("arn:aws:iam::%s:instance-profile/%s", i.accountID, name)
	}
	breakGlass := map[string]string{}
	for _, n := range st.BreakGlassRoleNames {
		breakGlass[n] = roleARN(n)
	}
	customRoles := map[string]string{}
	for _, n := range st.CustomRoleNames {
		customRoles[n] = roleARN(n)
	}
	// Always non-nil — empty maps stringify to "map[]" which the StringToMap
	// decode hook handles cleanly. Nil maps land as NULL in hstore and break
	// downstream decode with "expected a map, got 'string'".
	installInputs := map[string]string{}
	if i.cfg != nil && len(i.cfg.InstallInputs) > 0 {
		installInputs = i.cfg.InstallInputs
	}
	data := map[string]any{
		"request_type":             "Create",
		"phone_home_type":          "aws",
		"account_id":               i.accountID,
		"region":                   i.opts.AWSRegion,
		"vpc_id":                   st.VPCID,
		"runner_subnet":            st.RunnerSubnetID,
		"public_subnets":           strings.Join(st.PublicSubnetIDs, ","),
		"private_subnets":          strings.Join(st.PrivateSubnetIDs, ","),
		"runner_security_group_id": st.RunnerSecurityGroupID,
		"runner_iam_role_arn":      roleARN(st.RunnerRoleName),
		"runner_instance_profile":  instanceProfileARN(st.RunnerInstanceProfileName),
		"runner_asg_name":          st.RunnerASGName,
		"runner_log_group_name":    st.RunnerLogGroupName,
		"provision_iam_role_arn":   roleARN(st.ProvisionRoleName),
		"maintenance_iam_role_arn": roleARN(st.MaintenanceRoleName),
		"deprovision_iam_role_arn": roleARN(st.DeprovisionRoleName),
	}
	// Always emit map-typed keys, even when empty. Customer dashboard
	// templates reference `.nuon.install_stack.outputs.break_glass_role_arns`
	// directly and explode if the key is missing. Empty Go maps stringify to
	// "map[]" via fmt.Sprintf("%v", v); the StringToMapDecodeHook handles
	// that input cleanly. The previous omit-when-empty workaround was for nil
	// maps landing as NULL in hstore — non-nil empty maps don't have that
	// problem.
	data["break_glass_role_arns"] = breakGlass
	data["custom_role_arns"] = customRoles
	data["install_inputs"] = installInputs
	for k, v := range st.SecretARNs {
		data[k] = v
	}
	return data
}

// Deprovision tears down everything in the state file in reverse order.
func (i *Installer) Deprovision(ctx context.Context) error {
	st, err := stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return err
	}
	// Hydrate config — Deprovision needs cfg to know which secret names to
	// delete. If StackRun isn't configured, fall back to an empty config; the
	// stack package's Delete* funcs all tolerate a sparse config.
	if i.opts.StackRun != nil {
		runClient := stackrun.New(stackrun.Config{
			CtlAPIURL:   i.opts.StackRun.CtlAPIURL,
			InstallID:   i.opts.InstallID,
			PhoneHomeID: i.opts.StackRun.PhoneHomeID,
		})
		resp, err := runClient.CreateRun(ctx, stackrun.RunKindDeprovision)
		if err != nil {
			i.log.Warn("create deprovision run (continuing with empty config)", "err", err.Error())
			i.cfg = &stack.Config{InstallID: i.opts.InstallID}
		} else if resp.Config != nil {
			i.cfg = resp.Config
			i.cfg.InstallID = i.opts.InstallID
		} else {
			i.cfg = &stack.Config{InstallID: i.opts.InstallID}
		}
	} else {
		i.cfg = &stack.Config{InstallID: i.opts.InstallID}
	}
	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *stack.State) error
	}{
		{"delete-runner-compute", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteRunnerCompute(ctx, l, i.ec2c, i.asgc, i.logsc, s)
		}},
		{"delete-secrets", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteSecrets(ctx, l, i.smc, s, i.cfg)
		}},
		{"delete-iam", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteIAMRoles(ctx, l, i.iamc, s)
		}},
		{"delete-vpc", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteVPC(ctx, l, i.ec2c, s)
		}},
	}
	for _, s := range steps {
		log := i.log.With("step", s.name)
		log.Info("step starting")
		if err := s.run(ctx, log, st); err != nil {
			log.Error("step failed", "err", err.Error())
			return fmt.Errorf("step %s: %w", s.name, err)
		}
		log.Info("step completed")
	}
	if err := st.Delete(); err != nil {
		return fmt.Errorf("delete state file: %w", err)
	}
	i.log.Info("deprovision complete")
	return nil
}

// Status returns the current persisted state.
func (i *Installer) Status() (*stack.State, error) {
	return stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
}

func (i *Installer) stepValidateAWS(ctx context.Context, log *slog.Logger, _ *stack.State) error {
	out, err := i.stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("sts get-caller-identity: %w", err)
	}
	i.accountID = aws.ToString(out.Account)
	log.Info("aws caller identity",
		"account", i.accountID,
		"arn", aws.ToString(out.Arn),
	)
	return nil
}
