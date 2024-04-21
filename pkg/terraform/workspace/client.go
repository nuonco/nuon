package workspace

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock_test.go -source=client.go -package=workspace

func (w *workspace) getClient(ctx context.Context, log hclog.Logger) (Terraform, error) {
	tf, err := tfexec.NewTerraform(w.root, w.execPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get terraform client: %w", err)
	}

	if err := tf.SetEnv(w.envVars); err != nil {
		return nil, fmt.Errorf("unable to set environment variables: %w", err)
	}

	tf.SetLogger(log.StandardLogger(nil))
	return tf, nil
}

// Terraform represents the client that terraform exposes in `tfexec`. We interface it here, so we can mock it and
// expose tests for this locally.
type Terraform interface {
	// SetEnv allows you to override environment variables, this should not be used for any well known
	// Terraform environment variables that are already covered in options. Pass nil to copy the values
	// from os.Environ. Attempting to set environment variables that should be managed manually will
	// result in ErrManualEnvVar being returned.
	SetEnv(env map[string]string) error
	// SetLogger specifies a logger for tfexec to use.
	// SetStdout specifies a writer to stream stdout to for every command.
	//
	// This should be used for information or logging purposes only, not control
	// flow. Any parsing necessary should be added as functionality to this package.
	SetStdout(w io.Writer)
	// SetStderr specifies a writer to stream stderr to for every command.
	//
	// This should be used for information or logging purposes only, not control
	// flow. Any parsing necessary should be added as functionality to this package.
	SetStderr(w io.Writer)
	// SetLog sets the TF_LOG environment variable for Terraform CLI execution.
	// This must be combined with a call to SetLogPath to take effect.
	//
	// This is only compatible with Terraform CLI 0.15.0 or later as setting the
	// log level was unreliable in earlier versions. It will default to TRACE when
	// SetLogPath is called on versions 0.14.11 and earlier, or if SetLogCore and
	// SetLogProvider have not been called before SetLogPath on versions 0.15.0 and
	// later.
	SetLog(log string) error
	// SetLogCore sets the TF_LOG_CORE environment variable for Terraform CLI
	// execution. This must be combined with a call to SetLogPath to take effect.
	//
	// This is only compatible with Terraform CLI 0.15.0 or later.
	SetLogCore(logCore string) error
	// SetLogPath sets the TF_LOG_PATH environment variable for Terraform CLI
	// execution.
	SetLogPath(path string) error
	// SetLogProvider sets the TF_LOG_PROVIDER environment variable for Terraform
	// CLI execution. This must be combined with a call to SetLogPath to take
	// effect.
	//
	// This is only compatible with Terraform CLI 0.15.0 or later.
	SetLogProvider(logProvider string) error
	// SetAppendUserAgent sets the TF_APPEND_USER_AGENT environment variable for
	// Terraform CLI execution.
	SetAppendUserAgent(ua string) error
	// SetDisablePluginTLS sets the TF_DISABLE_PLUGIN_TLS environment variable for
	// Terraform CLI execution.
	SetDisablePluginTLS(disabled bool) error
	// SetSkipProviderVerify sets the TF_SKIP_PROVIDER_VERIFY environment variable
	// for Terraform CLI execution. This is no longer used in 0.13.0 and greater.
	SetSkipProviderVerify(skip bool) error
	// WorkingDir returns the working directory for Terraform.
	WorkingDir() string
	// ExecPath returns the path to the Terraform executable.
	ExecPath() string
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	// Apply represents the terraform apply subcommand.
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	// ApplyJSON represents the terraform apply subcommand with the `-json` flag.
	// Using the `-json` flag will result in
	// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
	// JSON being written to the supplied `io.Writer`. ApplyJSON is likely to be
	// removed in a future major version in favour of Apply returning JSON by default.
	ApplyJSON(ctx context.Context, w io.Writer, opts ...tfexec.ApplyOption) error
	// Destroy represents the terraform destroy subcommand.
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	// DestroyJSON represents the terraform destroy subcommand with the `-json` flag.
	// Using the `-json` flag will result in
	// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
	// JSON being written to the supplied `io.Writer`. DestroyJSON is likely to be
	// removed in a future major version in favour of Destroy returning JSON by default.
	DestroyJSON(ctx context.Context, w io.Writer, opts ...tfexec.DestroyOption) error
	// FormatString formats a passed string.
	FormatString(ctx context.Context, content string) (string, error)
	// Format performs formatting on the unformatted io.Reader (as stdin to the CLI) and returns
	// the formatted result on the formatted io.Writer.
	Format(ctx context.Context, unformatted io.Reader, formatted io.Writer) error
	// FormatWrite attempts to format and modify all config files in the working or selected (via DirOption) directory.
	FormatWrite(ctx context.Context, opts ...tfexec.FormatOption) error
	// FormatCheck returns true if the config files in the working or selected (via DirOption) directory are already formatted.
	FormatCheck(ctx context.Context, opts ...tfexec.FormatOption) (bool, []string, error)
	ForceUnlock(ctx context.Context, lockID string, opts ...tfexec.ForceUnlockOption) error
	Graph(ctx context.Context, opts ...tfexec.GraphOption) (string, error)
	// Output represents the terraform output subcommand.
	Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
	Validate(ctx context.Context) (*tfjson.ValidateOutput, error)
	// Plan executes `terraform plan` with the specified options and waits for it
	// to complete.
	//
	// The returned boolean is false when the plan diff is empty (no changes) and
	// true when the plan diff is non-empty (changes present).
	//
	// The returned error is nil if `terraform plan` has been executed and exits
	// with either 0 or 2.
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	// PlanJSON executes `terraform plan` with the specified options as well as the
	// `-json` flag and waits for it to complete.
	//
	// Using the `-json` flag will result in
	// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
	// JSON being written to the supplied `io.Writer`.
	//
	// The returned boolean is false when the plan diff is empty (no changes) and
	// true when the plan diff is non-empty (changes present).
	//
	// The returned error is nil if `terraform plan` has been executed and exits
	// with either 0 or 2.
	//
	// PlanJSON is likely to be removed in a future major version in favour of
	// Plan returning JSON by default.
	PlanJSON(ctx context.Context, w io.Writer, opts ...tfexec.PlanOption) (bool, error)
	// ProvidersLock represents the `terraform providers lock` command
	ProvidersLock(ctx context.Context, opts ...tfexec.ProvidersLockOption) error
	Get(ctx context.Context, opts ...tfexec.GetCmdOption) error
	// ProvidersSchema represents the terraform providers schema -json subcommand.
	ProvidersSchema(ctx context.Context) (*tfjson.ProviderSchemas, error)
	// Refresh represents the terraform refresh subcommand.
	Refresh(ctx context.Context, opts ...tfexec.RefreshCmdOption) error
	// RefreshJSON represents the terraform refresh subcommand with the `-json` flag.
	// Using the `-json` flag will result in
	// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
	// JSON being written to the supplied `io.Writer`. RefreshJSON is likely to be
	// removed in a future major version in favour of Refresh returning JSON by default.
	RefreshJSON(ctx context.Context, w io.Writer, opts ...tfexec.RefreshCmdOption) error
	// Show reads the default state path and outputs the state.
	// To read a state or plan file, ShowState or ShowPlan must be used instead.
	Show(ctx context.Context, opts ...tfexec.ShowOption) (*tfjson.State, error)
	// ShowStateFile reads a given state file and outputs the state.
	ShowStateFile(ctx context.Context, statePath string, opts ...tfexec.ShowOption) (*tfjson.State, error)
	// ShowPlanFile reads a given plan file and outputs the plan.
	ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error)
	// ShowPlanFileRaw reads a given plan file and outputs the plan in a
	// human-friendly, opaque format.
	ShowPlanFileRaw(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (string, error)
}
