// Package stack is the Nuon Stack SDK. It provisions and tears down the
// resources that make up a Nuon install stack (VPC + subnets, IAM roles,
// Secrets Manager entries, runner EC2 ASG) and reports run status back to
// ctl-api over the public phone-home endpoint.
//
// The actual resource lifecycle is delegated to a provisioning method
// (see internal/core.Provisioner): the AWS SDK implementation lives in
// internal/awssdk, with Terraform and CloudFormation methods to follow. This
// package owns the cross-cutting concerns — run reporting, log-stream wiring,
// and config hydration — and selects which method to drive.
//
// Customer-facing clients (stack-cli, embedded Go consumers) construct an
// Installer with FromURL when bootstrapping from a dashboard-rendered URL,
// or with New for offline state inspection.
package stack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nuonco/nuon/sdks/stack/internal/awssdk"
	"github.com/nuonco/nuon/sdks/stack/internal/core"
	"github.com/nuonco/nuon/sdks/stack/internal/logstream"
	"github.com/nuonco/nuon/sdks/stack/internal/terraform"
)

// Outputs is the method-agnostic result of a provision: the fully-resolved
// values that make up the phone-home payload. Re-exported from internal/core.
type Outputs = core.Outputs

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

	return &Installer{
		opts:   opts,
		log:    prov.Logger().With("install_id", opts.InstallID, "aws_region", opts.AWSRegion),
		sysLog: prov.SystemLogger().With("install_id", opts.InstallID, "aws_region", opts.AWSRegion),
		prov:   prov,
	}, nil
}

// selectProvisioner constructs the provisioning method for this run. The
// method is resolved from Options (explicit CLI override) first, then the
// ctl-api Config, then the default. Construction is deferred until the Config
// is hydrated because ctl-api decides the method per install.
func (i *Installer) selectProvisioner(ctx context.Context, cfg *Config) (core.Provisioner, error) {
	method := i.opts.Method
	if method == "" {
		method = cfg.Method
	}
	if method == "" {
		method = core.DefaultMethod
	}
	switch method {
	case core.MethodTerraform:
		return terraform.New(i.opts.AWSRegion), nil
	case core.MethodAWSSDK:
		return awssdk.New(ctx, i.opts.AWSRegion)
	default:
		return nil, fmt.Errorf("unknown provisioning method %q", method)
	}
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
				i.sysLog.Warn("init otlp log stream failed (continuing with prior logger)", "err", err.Error())
			} else {
				if i.prov != nil {
					_ = i.prov.Shutdown(ctx)
				}
				i.prov = otelProv
				i.log = otelProv.Logger().With(
					"install_id", i.opts.InstallID,
					"aws_region", i.opts.AWSRegion,
				)
				i.sysLog = otelProv.SystemLogger().With(
					"install_id", i.opts.InstallID,
					"aws_region", i.opts.AWSRegion,
				)
			}
		}
	} else {
		// No stackRun configured: run with an empty config so the method can
		// still be exercised end-to-end.
		i.cfg = &Config{InstallID: i.opts.InstallID}
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
	if out == nil {
		// Failure path: ctl-api skips output processing for failed runs, so a
		// minimal payload is sufficient.
		return data
	}

	data["account_id"] = out.AccountID
	data["region"] = out.Region
	data["vpc_id"] = out.VPCID
	data["runner_subnet"] = out.RunnerSubnetID
	data["public_subnets"] = strings.Join(out.PublicSubnetIDs, ",")
	data["private_subnets"] = strings.Join(out.PrivateSubnetIDs, ",")
	data["runner_security_group_id"] = out.RunnerSecurityGroupID
	data["runner_iam_role_arn"] = out.RunnerIAMRoleARN
	data["runner_instance_profile"] = out.RunnerInstanceProfileARN
	data["runner_asg_name"] = out.RunnerASGName
	data["runner_log_group_name"] = out.RunnerLogGroupName
	data["provision_iam_role_arn"] = out.ProvisionRoleARN
	data["maintenance_iam_role_arn"] = out.MaintenanceRoleARN
	data["deprovision_iam_role_arn"] = out.DeprovisionRoleARN

	// Always emit map-typed keys, even when empty. Customer dashboard
	// templates reference `.nuon.install_stack.outputs.break_glass_role_arns`
	// directly and explode if the key is missing. Empty Go maps stringify to
	// "map[]" via fmt.Sprintf("%v", v); the StringToMapDecodeHook handles
	// that input cleanly.
	breakGlass := out.BreakGlassRoleARNs
	if breakGlass == nil {
		breakGlass = map[string]string{}
	}
	customRoles := out.CustomRoleARNs
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
	for k, v := range out.SecretARNs {
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

	provisioner, err := i.selectProvisioner(ctx, i.cfg)
	if err != nil {
		return err
	}
	return provisioner.Deprovision(ctx, i.log, i.sysLog, i.cfg)
}

// Status returns the current persisted outputs for the install.
func (i *Installer) Status(ctx context.Context) (*Outputs, error) {
	cfg := &Config{InstallID: i.opts.InstallID, AWSRegion: i.opts.AWSRegion, Method: i.opts.Method}
	provisioner, err := i.selectProvisioner(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return provisioner.Status(ctx, cfg)
}
