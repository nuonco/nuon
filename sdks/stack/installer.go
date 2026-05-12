// Package stack is the Nuon Stack SDK. It provisions and tears down the AWS
// resources that make up a Nuon install stack (VPC + subnets, IAM roles,
// Secrets Manager entries, runner EC2 ASG) and reports run status back to
// ctl-api over the public phone-home endpoint.
//
// Customer-facing clients (installer-cli, embedded Go consumers) construct an
// Installer with FromURL when bootstrapping from a dashboard-rendered URL,
// or with New for offline state inspection.
package stack

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

	"github.com/nuonco/nuon/sdks/stack/internal/logstream"
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

	// cfg is hydrated from the createRun response and threaded into every
	// resource provisioner. nil until Provision/Deprovision fetches it.
	cfg *Config

	// accountID is captured during stepValidateAWS so reportRun can build
	// IAM role ARNs without a second sts call on the failure path.
	accountID string

	// preCreatedRun, when set, makes Run/Deprovision skip their own createRun
	// call and use this response instead. Set by FromURL — the URL flow
	// calls createRun before we even know InstallID/AWSRegion, so by the
	// time we get to Run() the response is already in hand.
	preCreatedRun *createRunResponse
	runClient     *runClient
}

// FromURL POSTs to the create-run URL the dashboard renders, then bootstraps
// an Installer from the response. The CLI's single-argument flow uses this —
// the caller doesn't need to know install_id, region, or runner details up
// front; the API hands them back in the response.
//
// Kind drives the kind path segment on the POST and propagates into the run
// row in ctl-api.
//
// On success the returned Installer already has the createRun response
// cached, so the next call to Run/Deprovision will not double-create.
func FromURL(ctx context.Context, in URLOptions) (*Installer, error) {
	base, phoneHomeID, err := parseURL(in.URL)
	if err != nil {
		return nil, err
	}
	if in.Kind == "" {
		return nil, fmt.Errorf("URLOptions.Kind required")
	}

	client := newRunClient(runClientConfig{
		CtlAPIURL:   base,
		PhoneHomeID: phoneHomeID,
	})
	resp, err := client.createRun(ctx, in.Kind)
	if err != nil {
		return nil, fmt.Errorf("create stack run: %w", err)
	}
	if resp.Config == nil {
		return nil, fmt.Errorf("create stack run: ctl-api returned no config block")
	}
	if resp.Config.InstallID == "" || resp.Config.AWSRegion == "" {
		return nil, fmt.Errorf("create stack run: config missing install_id or aws_region")
	}

	opts := Options{
		InstallID: resp.Config.InstallID,
		AWSRegion: resp.Config.AWSRegion,
		stackRun: &stackRunConfig{
			CtlAPIURL:   base,
			PhoneHomeID: phoneHomeID,
		},
	}
	if resp.LogStream != nil && resp.LogStream.ID != "" {
		opts.logStream = &logStreamConfig{
			ID:           resp.LogStream.ID,
			WriteToken:   resp.LogStream.WriteToken,
			RunnerAPIURL: resp.LogStream.RunnerAPIURL,
		}
	}
	inst, err := New(ctx, opts)
	if err != nil {
		return nil, err
	}
	inst.preCreatedRun = resp
	inst.runClient = client
	return inst, nil
}

// New builds an Installer from explicit Options. Use this for offline state
// inspection (Status); use FromURL for actual provisioning. Caller must call
// Close to flush logs.
func New(ctx context.Context, opts Options) (*Installer, error) {
	if opts.InstallID == "" {
		return nil, fmt.Errorf("InstallID required")
	}
	if opts.AWSRegion == "" {
		return nil, fmt.Errorf("AWSRegion required")
	}

	var prov *logstream.Provider
	if opts.logStream == nil {
		prov = logstream.NewStdout("stack")
	} else {
		var err error
		prov, err = logstream.New(ctx, logstream.Config{
			RunnerAPIURL: opts.logStream.RunnerAPIURL,
			LogStreamID:  opts.logStream.ID,
			WriteToken:   opts.logStream.WriteToken,
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

// PreparedConfig returns the rendered Config attached to the pre-created run
// returned by FromURL. It's only populated between FromURL and the first
// Provision/Reprovision/Deprovision call — once the run starts, the cached
// response is consumed and this returns nil. Callers that want to preview
// the install config before kicking off provisioning use this.
func (i *Installer) PreparedConfig() *Config {
	if i.preCreatedRun == nil {
		return nil
	}
	return i.preCreatedRun.Config
}

func (i *Installer) Close(ctx context.Context) error {
	if i.prov != nil {
		return i.prov.Shutdown(ctx)
	}
	return nil
}

// Provision runs the full provisioning sequence. State is persisted to disk
// after each successful step so a partial failure can be cleaned up by Deprovision.
func (i *Installer) Provision(ctx context.Context) (*State, error) {
	return i.run(ctx, KindProvision)
}

// Reprovision is a re-run on an existing install. Functionally identical to
// Provision (both code paths are idempotent — every step is discover-or-create
// keyed on the install_id tag) but recorded as a distinct run kind so the
// dashboard can show first-time vs reconcile in the audit trail.
func (i *Installer) Reprovision(ctx context.Context) (*State, error) {
	return i.run(ctx, KindReprovision)
}

// run executes a provision-shaped workflow under the given kind.
func (i *Installer) run(ctx context.Context, kind Kind) (*State, error) {
	st, err := loadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return nil, err
	}

	// Report to ctl-api as a stack run. Required when configured: the response
	// also carries the OTLP log-stream credentials and the rendered Config we
	// need for visibility and resource provisioning.
	var rc *runClient
	var runID string
	if i.opts.stackRun != nil {
		rc = i.runClient
		if rc == nil {
			rc = newRunClient(runClientConfig{
				CtlAPIURL:   i.opts.stackRun.CtlAPIURL,
				PhoneHomeID: i.opts.stackRun.PhoneHomeID,
			})
		}
		// FromURL already issued createRun before we knew install_id /
		// region; reuse that response instead of double-creating.
		var resp *createRunResponse
		if i.preCreatedRun != nil {
			resp = i.preCreatedRun
			i.preCreatedRun = nil
		} else {
			var err error
			resp, err = rc.createRun(ctx, kind)
			if err != nil {
				return st, fmt.Errorf("create stack run: %w", err)
			}
		}
		runID = resp.ID
		i.log.Info("created stack run", "run_id", runID)
		if resp.Config == nil {
			return st, fmt.Errorf("create stack run: ctl-api returned no config block")
		}
		i.cfg = resp.Config
		i.cfg.InstallID = i.opts.InstallID
		// Propagate identity + cluster name into State so resource tagging
		// callsites can read them without holding cfg.
		st.OrgID = i.cfg.OrgID
		st.AppID = i.cfg.AppID
		st.ClusterName = i.cfg.ClusterName
		if st.ClusterName == "" {
			st.ClusterName = i.opts.InstallID
		}
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
		// No stackRun configured: run with an empty config so VPC/log-group/
		// secrets can still be exercised end-to-end.
		i.cfg = &Config{InstallID: i.opts.InstallID}
	}

	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *State) error
	}{
		{"validate-aws", i.stepValidateAWS},
		{"create-vpc", func(ctx context.Context, l *slog.Logger, s *State) error {
			return createVPC(ctx, l, i.ec2c, s)
		}},
		{"create-iam", func(ctx context.Context, l *slog.Logger, s *State) error {
			return createIAMRoles(ctx, l, i.iamc, i.stsc, s, i.cfg)
		}},
		{"create-secrets", func(ctx context.Context, l *slog.Logger, s *State) error {
			return ensureSecrets(ctx, l, i.smc, s, i.cfg)
		}},
		{"create-runner-compute", func(ctx context.Context, l *slog.Logger, s *State) error {
			// Cycle the running instance on every run — initial provision is a
			// no-op (no instance exists yet); reprovision picks up refreshed
			// tags / AMI / user-data without manual intervention.
			refresh := kind == KindProvision || kind == KindReprovision
			return ensureRunnerCompute(ctx, l, i.ec2c, i.iamc, i.asgc, i.logsc, s, i.cfg, refresh)
		}},
	}

	for _, s := range steps {
		log := i.log.With("step", s.name)
		log.Info("step starting")
		if err := s.run(ctx, log, st); err != nil {
			log.Error("step failed", "err", err.Error())
			i.reportRun(ctx, rc, runID, "failed", err.Error(), st)
			return st, fmt.Errorf("step %s: %w", s.name, err)
		}
		log.Info("step completed")
	}

	i.log.Info("provision complete", "runner_asg", st.RunnerASGName)
	i.reportRun(ctx, rc, runID, "succeeded", "", st)
	return st, nil
}

// reportRun is best-effort; failures only log. Builds the phone-home payload
// described in install-stacks/aws/phone_home.tf so app templates resolving
// `nuon.install_stack.outputs.*` see identical key sets across CFN/TF/SDK.
func (i *Installer) reportRun(ctx context.Context, c *runClient, runID, status, statusDesc string, st *State) {
	if c == nil || runID == "" {
		return
	}
	data := i.buildPhoneHomePayload(st)
	if err := c.updateRun(ctx, runID, updateRunRequest{
		Status:            status,
		StatusDescription: statusDesc,
		Data:              data,
	}); err != nil {
		i.log.Warn("update stack run failed", "err", err.Error(), "status", status)
	}
}

func (i *Installer) buildPhoneHomePayload(st *State) map[string]any {
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
	// that input cleanly.
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
	st, err := loadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return err
	}
	// Hydrate config — Deprovision needs cfg to know which secret names to
	// delete. If stackRun isn't configured, fall back to an empty config; the
	// Delete* funcs all tolerate a sparse config.
	if i.opts.stackRun != nil {
		rc := i.runClient
		if rc == nil {
			rc = newRunClient(runClientConfig{
				CtlAPIURL:   i.opts.stackRun.CtlAPIURL,
				PhoneHomeID: i.opts.stackRun.PhoneHomeID,
			})
		}
		var resp *createRunResponse
		var err error
		if i.preCreatedRun != nil {
			resp = i.preCreatedRun
			i.preCreatedRun = nil
		} else {
			resp, err = rc.createRun(ctx, KindDeprovision)
		}
		if err != nil {
			i.log.Warn("create deprovision run (continuing with empty config)", "err", err.Error())
			i.cfg = &Config{InstallID: i.opts.InstallID}
		} else if resp.Config != nil {
			i.cfg = resp.Config
			i.cfg.InstallID = i.opts.InstallID
		} else {
			i.cfg = &Config{InstallID: i.opts.InstallID}
		}
	} else {
		i.cfg = &Config{InstallID: i.opts.InstallID}
	}
	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *State) error
	}{
		{"delete-runner-compute", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteRunnerCompute(ctx, l, i.ec2c, i.asgc, i.logsc, s)
		}},
		{"delete-secrets", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteSecrets(ctx, l, i.smc, s, i.cfg)
		}},
		{"delete-iam", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteIAMRoles(ctx, l, i.iamc, s)
		}},
		{"delete-vpc", func(ctx context.Context, l *slog.Logger, s *State) error {
			return deleteVPC(ctx, l, i.ec2c, s)
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
func (i *Installer) Status() (*State, error) {
	return loadState(i.opts.InstallID, i.opts.AWSRegion)
}

func (i *Installer) stepValidateAWS(ctx context.Context, log *slog.Logger, _ *State) error {
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
