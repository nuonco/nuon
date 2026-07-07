// Package stack is the Nuon Stack SDK. It provisions and tears down the
// resources that make up a Nuon install stack (VPC + subnets, IAM roles,
// Secrets Manager entries, runner EC2 ASG) and reports run status back to
// ctl-api over the public phone-home endpoint.
//
// The actual resource lifecycle is delegated to the Terraform provisioner
// (see internal/core.Provisioner and internal/terraform). This package owns
// the cross-cutting concerns — run reporting, log-stream wiring, and config
// hydration.
//
// Customer-facing clients (stack-cli, embedded Go consumers) construct an
// Installer with FromURL when bootstrapping from a dashboard-rendered URL,
// or with New for offline state inspection.
package stack

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
	"github.com/nuonco/nuon/sdks/stack/internal/logstream"
	"github.com/nuonco/nuon/sdks/stack/internal/terraform"
)

// Outputs is the method-agnostic result of a provision: the fully-resolved
// values that make up the phone-home payload. Re-exported from internal/core.
type Outputs = core.Outputs

// AWSOutputs and GCPOutputs are the cloud-specific resolved values on Outputs,
// re-exported so consumers (e.g. the Terraform provider) can inspect and
// construct them without reaching into internal packages.
type (
	AWSOutputs = core.AWSOutputs
	GCPOutputs = core.GCPOutputs
)

// Installer provisions and tears down an install stack in a customer cloud
// account. It drives a single provisioning method selected at construction.
type Installer struct {
	opts Options
	// log emits under OTEL scope "oteljob" — user-visible job output the
	// dashboard surfaces by default (resource progress, step status, run
	// completion). sysLog emits under scope "system" for internal chatter
	// (transient retries, best-effort phone-home failures) that the dashboard
	// hides unless the user toggles "Include system logs".
	log    *slog.Logger
	sysLog *slog.Logger
	prov   *logstream.Provider

	// cfg is hydrated from the createRun response and threaded into the
	// provisioner. nil until Provision/Deprovision fetches it.
	cfg *Config

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
	if resp.Config.InstallID == "" {
		return nil, fmt.Errorf("create stack run: config missing install_id")
	}

	// Resolve the cloud: explicit override, then the ctl-api Config, then the
	// default. The location inputs come from different places per cloud — AWS
	// region from the Config, GCP project/region from the caller (URLOptions).
	cloud := in.Cloud
	if cloud == "" {
		cloud = resp.Config.Cloud
	}
	if cloud == "" {
		cloud = core.DefaultCloud
	}

	opts := Options{
		InstallID:         resp.Config.InstallID,
		Cloud:             cloud,
		Method:            in.Method,
		GCP:               in.GCP,
		InstallInputs:     in.InstallInputs,
		Secrets:           in.Secrets,
		Backend:           in.Backend,
		WorkDir:           in.WorkDir,
		TerraformExecPath: in.TerraformExecPath,
		stackRun: &stackRunConfig{
			CtlAPIURL:   base,
			PhoneHomeID: phoneHomeID,
		},
	}
	switch cloud {
	case core.CloudAWS:
		if resp.Config.AWS != nil {
			opts.AWSRegion = resp.Config.AWS.Region
		}
		if opts.AWSRegion == "" {
			return nil, fmt.Errorf("create stack run: config missing aws region")
		}
	case core.CloudGCP:
		// Project + region are customer-supplied at provision time — they are
		// not known server-side. They are collected by the interactive wizard
		// (or via URLOptions.GCP) after construction and validated when the run
		// starts, so we don't require them here.
	default:
		return nil, fmt.Errorf("unsupported cloud %q", cloud)
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

// FetchConfig reads the rendered install-stack configuration for the stack
// version identified by the create-run URL, without creating a run or mutating
// any state. It is the read-only counterpart to FromURL: callers that only need
// the config (e.g. the Terraform provider's nuon_stack data source, which feeds
// it to an install-stacks module in place of tfvars) use this instead of
// provisioning. The URL is the same /v1/stack-runs/{phone_home_id} form FromURL
// accepts.
func FetchConfig(ctx context.Context, url string) (*Config, error) {
	base, phoneHomeID, err := parseURL(url)
	if err != nil {
		return nil, err
	}
	client := newRunClient(runClientConfig{
		CtlAPIURL:   base,
		PhoneHomeID: phoneHomeID,
	})
	cfg, err := client.fetchConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("fetch stack config: ctl-api returned no config block")
	}
	return cfg, nil
}

// New builds an Installer from explicit Options. Use this for offline state
// inspection (Status); use FromURL for actual provisioning. Caller must call
// Close to flush logs.
func New(ctx context.Context, opts Options) (*Installer, error) {
	if opts.InstallID == "" {
		return nil, fmt.Errorf("InstallID required")
	}
	locKey, locVal, err := validateLocation(opts)
	if err != nil {
		return nil, err
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
				locKey:       locVal,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return &Installer{
		opts:   opts,
		log:    prov.Logger().With("install_id", opts.InstallID, locKey, locVal),
		sysLog: prov.SystemLogger().With("install_id", opts.InstallID, locKey, locVal),
		prov:   prov,
	}, nil
}

// validateLocation checks the caller-supplied location inputs for the resolved
// cloud and returns the log attribute (key, value) naming that location. AWS
// needs a region; GCP needs a project and region.
func validateLocation(opts Options) (string, string, error) {
	cloud := opts.Cloud
	if cloud == "" {
		cloud = core.DefaultCloud
	}
	switch cloud {
	case core.CloudAWS:
		if opts.AWSRegion == "" {
			return "", "", fmt.Errorf("AWSRegion required for cloud aws")
		}
		return "aws_region", opts.AWSRegion, nil
	case core.CloudGCP:
		// Location is collected interactively after construction; validated at
		// provision time, not here.
		return "gcp_region", opts.GCP.Region, nil
	default:
		return "", "", fmt.Errorf("unsupported cloud %q", cloud)
	}
}

// resolveCloud picks the target cloud: explicit Options override, then the
// hydrated Config, then the default.
func (i *Installer) resolveCloud() core.Cloud {
	if i.opts.Cloud != "" {
		return i.opts.Cloud
	}
	if i.cfg != nil && i.cfg.Cloud != "" {
		return i.cfg.Cloud
	}
	return core.DefaultCloud
}

// applyCloudOptions overlays the caller-supplied location/sizing inputs onto
// the hydrated Config. ctl-api provides the Nuon-generated fields; the customer
// supplies the rest (AWS region, GCP project/region/machine-type/GKE) via
// Options. It also pins Config.Cloud so downstream resolution is consistent.
func (i *Installer) applyCloudOptions() {
	if i.cfg == nil {
		return
	}
	cloud := i.resolveCloud()
	i.cfg.Cloud = cloud
	switch cloud {
	case core.CloudAWS:
		if i.cfg.AWS == nil {
			i.cfg.AWS = &core.AWSConfig{}
		}
		if i.cfg.AWS.Region == "" {
			i.cfg.AWS.Region = i.opts.AWSRegion
		}
	case core.CloudGCP:
		if i.cfg.GCP == nil {
			i.cfg.GCP = &core.GCPConfig{}
		}
		g, o := i.cfg.GCP, i.opts.GCP
		if o.ProjectID != "" {
			g.ProjectID = o.ProjectID
		}
		if o.Region != "" {
			g.Region = o.Region
		}
		if o.RunnerMachineType != "" {
			g.RunnerMachineType = o.RunnerMachineType
		}
		if o.HasGKENodePool != nil {
			g.HasGKENodePool = o.HasGKENodePool
		}
		if o.GKENodePoolSAEmail != "" {
			g.GKENodePoolSAEmail = o.GKENodePoolSAEmail
		}
	}
}

// applyTerraformOptions overlays the caller-supplied terraform-method controls
// (remote backend, work dir, pre-installed binary) onto the hydrated Config.
// ctl-api does not carry these — they are execution-environment concerns the
// caller (e.g. the Terraform provider) supplies. The backend key/prefix default
// to an install-scoped path, and the S3 region falls back to the AWS region.
func (i *Installer) applyTerraformOptions() {
	if i.cfg == nil {
		return
	}
	if i.opts.WorkDir != "" {
		i.cfg.TerraformWorkDir = i.opts.WorkDir
	}
	if i.opts.TerraformExecPath != "" {
		i.cfg.TerraformExecPath = i.opts.TerraformExecPath
	}
	if i.opts.Backend.Bucket == "" {
		return
	}
	be := i.opts.Backend
	switch i.resolveCloud() {
	case core.CloudAWS:
		if be.Key == "" {
			be.Key = fmt.Sprintf("nuon/%s/terraform.tfstate", i.opts.InstallID)
		}
		if be.Region == "" {
			be.Region = i.opts.AWSRegion
		}
	case core.CloudGCP:
		if be.Prefix == "" {
			be.Prefix = fmt.Sprintf("nuon/%s", i.opts.InstallID)
		}
	}
	i.cfg.TerraformBackend = &be
}

// overlayInputsAndSecrets applies caller-supplied install-input and secret
// values onto the hydrated Config. The Config from ctl-api declares which keys
// exist; the caller supplies values. Unknown keys are rejected, and a value for
// an auto-generated secret is rejected — the module generates those.
func (i *Installer) overlayInputsAndSecrets() error {
	if i.cfg == nil {
		return nil
	}

	if len(i.opts.InstallInputs) > 0 {
		known := make(map[string]struct{}, len(i.cfg.InstallInputs)+len(i.cfg.RequiredInputs))
		for k := range i.cfg.InstallInputs {
			known[k] = struct{}{}
		}
		for _, k := range i.cfg.RequiredInputs {
			known[k] = struct{}{}
		}
		if i.cfg.InstallInputs == nil {
			i.cfg.InstallInputs = map[string]string{}
		}
		for k, v := range i.opts.InstallInputs {
			if _, ok := known[k]; !ok {
				return fmt.Errorf("unknown install input %q", k)
			}
			i.cfg.InstallInputs[k] = v
		}
	}

	if len(i.opts.Secrets) > 0 {
		autoGen := make(map[string]struct{}, len(i.cfg.AutoGenerateSecrets))
		for _, name := range i.cfg.AutoGenerateSecrets {
			autoGen[name] = struct{}{}
		}
		for k, v := range i.opts.Secrets {
			if _, isAuto := autoGen[k]; isAuto {
				return fmt.Errorf("secret %q is auto-generated and cannot be set", k)
			}
			sec, ok := i.cfg.Secrets[k]
			if !ok {
				return fmt.Errorf("unknown secret %q", k)
			}
			sec.Value = v
			i.cfg.Secrets[k] = sec
		}
	}

	return nil
}

// validateProvisionConfig checks the resolved Config has the cloud's required
// location inputs before any resources change. For GCP these are collected
// interactively (or via options) after construction, so this is the first
// point they're guaranteed present.
func (i *Installer) validateProvisionConfig() error {
	switch i.resolveCloud() {
	case core.CloudGCP:
		if i.cfg.GCP == nil || i.cfg.GCP.ProjectID == "" || i.cfg.GCP.Region == "" {
			return fmt.Errorf("gcp install requires a project and region")
		}
	case core.CloudAWS:
		if i.cfg.AWS == nil || i.cfg.AWS.Region == "" {
			return fmt.Errorf("aws install requires a region")
		}
	}
	return nil
}

// validateRequiredValues checks that every required install input and secret
// has a non-empty value. Enforced only on provision/reprovision (not
// deprovision, where missing values shouldn't block teardown). Reports all
// missing fields at once so the user can fix them in a single pass.
func (i *Installer) validateRequiredValues() error {
	var missing []string
	for _, name := range i.cfg.RequiredInputs {
		if strings.TrimSpace(i.cfg.InstallInputs[name]) == "" {
			missing = append(missing, "input "+name)
		}
	}
	for name, sec := range i.cfg.Secrets {
		if sec.Required && strings.TrimSpace(sec.Value) == "" {
			missing = append(missing, "secret "+name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("required values not set: %s", strings.Join(missing, ", "))
	}
	return nil
}

// locationAttr returns the log attribute (key, value) naming the run's
// location for the resolved cloud.
func (i *Installer) locationAttr() (string, string) {
	if i.resolveCloud() == core.CloudGCP {
		return "gcp_region", i.opts.GCP.Region
	}
	return "aws_region", i.opts.AWSRegion
}

// selectProvisioner constructs the Terraform provisioner for this run. The
// target cloud is resolved from Options (explicit override), then the ctl-api
// Config, then the default. Construction is deferred until the Config is
// hydrated because ctl-api decides the cloud per install.
func (i *Installer) selectProvisioner(_ context.Context, cfg *Config) (core.Provisioner, error) {
	cloud := i.opts.Cloud
	if cloud == "" {
		cloud = cfg.Cloud
	}
	if cloud == "" {
		cloud = core.DefaultCloud
	}
	if err := core.ValidateCloud(cloud); err != nil {
		return nil, err
	}
	return terraform.New(cloud)
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

// Provision runs the full provisioning sequence via the selected method.
func (i *Installer) Provision(ctx context.Context) (*Outputs, error) {
	return i.run(ctx, KindProvision)
}

// Reprovision is a re-run on an existing install. Functionally identical to
// Provision (both code paths are idempotent — every step is discover-or-create
// keyed on the install_id tag) but recorded as a distinct run kind so the
// dashboard can show first-time vs reconcile in the audit trail.
func (i *Installer) Reprovision(ctx context.Context) (*Outputs, error) {
	return i.run(ctx, KindReprovision)
}

// run executes a provision-shaped workflow under the given kind: it reports
// the run to ctl-api (which mints the log-stream credentials and rendered
// config), then delegates the resource lifecycle to the provisioner.
func (i *Installer) run(ctx context.Context, kind Kind) (*Outputs, error) {
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
				return nil, fmt.Errorf("create stack run: %w", err)
			}
		}
		runID = resp.ID
		i.log.Info("created stack run", "run_id", runID)
		if resp.Config == nil {
			return nil, fmt.Errorf("create stack run: ctl-api returned no config block")
		}
		i.cfg = resp.Config
		i.cfg.InstallID = i.opts.InstallID
		// Overlay the caller-supplied cloud inputs (e.g. GCP project/region,
		// AWS region) the ctl-api Config doesn't carry.
		i.applyCloudOptions()
		i.applyTerraformOptions()
		// Validate the bootstrap fields before any resources change. The
		// runner's init script reads nuon_runner_id / nuon_runner_api_url
		// from EC2 instance tags via IMDSv2 — empty values would let
		// provisioning succeed but leave the runner unable to authenticate,
		// surfacing as a silent "never connects" later.
		if i.cfg.RunnerID == "" {
			return nil, fmt.Errorf("ctl-api config missing runner_id — install has no runner attached")
		}
		if i.cfg.RunnerAPIURL == "" {
			return nil, fmt.Errorf("ctl-api config missing runner_api_url — set RunnerGroupSettings.RunnerAPIURL on this install's runner group")
		}
		// Swap the stdout-only provider for an OTLP one fed by the credentials
		// ctl-api just minted for this run.
		if resp.LogStream != nil && resp.LogStream.ID != "" {
			locKey, locVal := i.locationAttr()
			otelProv, err := logstream.New(ctx, logstream.Config{
				RunnerAPIURL: resp.LogStream.RunnerAPIURL,
				LogStreamID:  resp.LogStream.ID,
				WriteToken:   resp.LogStream.WriteToken,
				ServiceName:  "stack",
				Attrs: map[string]string{
					"install_id": i.opts.InstallID,
					locKey:       locVal,
				},
			})
			if err != nil {
				i.sysLog.Warn("init otlp log stream failed (continuing with prior logger)", "err", err.Error())
			} else {
				if i.prov != nil {
					_ = i.prov.Shutdown(ctx)
				}
				i.prov = otelProv
				i.log = otelProv.Logger().With(
					"install_id", i.opts.InstallID,
					locKey, locVal,
				)
				i.sysLog = otelProv.SystemLogger().With(
					"install_id", i.opts.InstallID,
					locKey, locVal,
				)
			}
		}
	} else {
		// No stackRun configured: run with an empty config so the method can
		// still be exercised end-to-end.
		i.cfg = &Config{InstallID: i.opts.InstallID}
		i.applyCloudOptions()
		i.applyTerraformOptions()
	}

	if err := i.overlayInputsAndSecrets(); err != nil {
		i.reportRun(ctx, rc, runID, "failed", err.Error(), nil)
		return nil, err
	}
	if err := i.validateProvisionConfig(); err != nil {
		i.reportRun(ctx, rc, runID, "failed", err.Error(), nil)
		return nil, err
	}
	if err := i.validateRequiredValues(); err != nil {
		i.reportRun(ctx, rc, runID, "failed", err.Error(), nil)
		return nil, err
	}

	provisioner, err := i.selectProvisioner(ctx, i.cfg)
	if err != nil {
		i.reportRun(ctx, rc, runID, "failed", err.Error(), nil)
		return nil, err
	}

	// When the method reports its own run (e.g. the Terraform module's
	// phone-home), the SDK must not also report or it double-reports.
	ownReporting := provisioner.ReportsOwnRun()

	outputs, err := provisioner.Provision(ctx, i.log, i.sysLog, i.cfg, kind)
	if err != nil {
		if !ownReporting {
			i.reportRun(ctx, rc, runID, "failed", err.Error(), nil)
		}
		return nil, err
	}
	if !ownReporting {
		i.reportRun(ctx, rc, runID, "succeeded", "", outputs)
	}
	return outputs, nil
}

// reportRun is best-effort; failures only log. Builds the phone-home payload
// described in install-stacks/aws/phone_home.tf so app templates resolving
// `nuon.install_stack.outputs.*` see identical key sets across CFN/TF/SDK.
func (i *Installer) reportRun(ctx context.Context, c *runClient, runID, status, statusDesc string, out *Outputs) {
	if c == nil || runID == "" {
		return
	}
	if err := c.updateRun(ctx, runID, updateRunRequest{
		Status:            status,
		StatusDescription: statusDesc,
		Data:              i.buildPhoneHomePayload(out),
	}); err != nil {
		i.sysLog.Warn("update stack run failed", "err", err.Error(), "status", status)
	}
}

func (i *Installer) buildPhoneHomePayload(out *Outputs) map[string]any {
	data := map[string]any{
		"request_type":    "Create",
		"phone_home_type": "aws",
	}
	if out == nil || out.AWS == nil {
		// Failure path: ctl-api skips output processing for failed runs, so a
		// minimal payload is sufficient.
		return data
	}
	aws := out.AWS

	data["account_id"] = aws.AccountID
	data["region"] = aws.Region
	data["vpc_id"] = aws.VPCID
	data["runner_subnet"] = aws.RunnerSubnetID
	data["public_subnets"] = strings.Join(aws.PublicSubnetIDs, ",")
	data["private_subnets"] = strings.Join(aws.PrivateSubnetIDs, ",")
	data["runner_security_group_id"] = aws.RunnerSecurityGroupID
	data["runner_iam_role_arn"] = aws.RunnerIAMRoleARN
	data["runner_instance_profile"] = aws.RunnerInstanceProfileARN
	data["runner_asg_name"] = aws.RunnerASGName
	data["runner_log_group_name"] = aws.RunnerLogGroupName
	data["provision_iam_role_arn"] = aws.ProvisionRoleARN
	data["maintenance_iam_role_arn"] = aws.MaintenanceRoleARN
	data["deprovision_iam_role_arn"] = aws.DeprovisionRoleARN

	// Always emit map-typed keys, even when empty. Customer dashboard
	// templates reference `.nuon.install_stack.outputs.break_glass_role_arns`
	// directly and explode if the key is missing. Empty Go maps stringify to
	// "map[]" via fmt.Sprintf("%v", v); the StringToMapDecodeHook handles
	// that input cleanly.
	breakGlass := aws.BreakGlassRoleARNs
	if breakGlass == nil {
		breakGlass = map[string]string{}
	}
	customRoles := aws.CustomRoleARNs
	if customRoles == nil {
		customRoles = map[string]string{}
	}
	installInputs := out.InstallInputs
	if installInputs == nil {
		installInputs = map[string]string{}
	}
	data["break_glass_role_arns"] = breakGlass
	data["custom_role_arns"] = customRoles
	data["install_inputs"] = installInputs
	for k, v := range aws.SecretARNs {
		data[k] = v
	}
	return data
}

// Deprovision tears down everything the selected method created.
func (i *Installer) Deprovision(ctx context.Context) error {
	// Hydrate config — Deprovision needs cfg to know which secret names to
	// delete. If stackRun isn't configured, fall back to an empty config; the
	// method's delete path tolerates a sparse config.
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
			i.sysLog.Warn("create deprovision run (continuing with empty config)", "err", err.Error())
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
	i.applyCloudOptions()
	i.applyTerraformOptions()
	if err := i.validateProvisionConfig(); err != nil {
		return err
	}

	provisioner, err := i.selectProvisioner(ctx, i.cfg)
	if err != nil {
		return err
	}
	return provisioner.Deprovision(ctx, i.log, i.sysLog, i.cfg)
}

// Status returns the current persisted outputs for the install.
func (i *Installer) Status(ctx context.Context) (*Outputs, error) {
	i.cfg = &Config{InstallID: i.opts.InstallID, Method: i.opts.Method}
	i.applyCloudOptions()
	i.applyTerraformOptions()
	provisioner, err := i.selectProvisioner(ctx, i.cfg)
	if err != nil {
		return nil, err
	}
	return provisioner.Status(ctx, i.cfg)
}
