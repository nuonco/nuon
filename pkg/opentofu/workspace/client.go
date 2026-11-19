package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock_test.go -source=client.go -package=workspace

// Tofu is the client this package uses to execute the tofu CLI. It mirrors the
// subset of tfexec's Terraform interface that the terraform workspace uses, so
// workspace code (and its tests) read like pkg/terraform/workspace's — but it
// is implemented with exec.Command against the tofu binary, never with tfexec.
//
// See the package comment in workspace.go for why tfexec cannot drive tofu.
type Tofu interface {
	// Version returns the parsed OpenTofu version. The client caches the
	// result; pass skipCache to force a fresh `tofu version` run.
	Version(ctx context.Context, skipCache bool) (*goversion.Version, error)
	// Init represents the tofu init subcommand, pointed at the workspace's
	// backend config file.
	Init(ctx context.Context, backendConfig string) error
	// ApplyJSON represents the tofu apply subcommand with the `-json` flag.
	// Using the `-json` flag results in machine-readable JSON being written to
	// the supplied io.Writer, mirroring how tfexec's ApplyJSON streams
	// terraform's JSON output.
	ApplyJSON(ctx context.Context, w io.Writer, opts ...ApplyOption) error
	// DestroyJSON represents the tofu destroy subcommand with the `-json` flag.
	DestroyJSON(ctx context.Context, w io.Writer, opts ...DestroyOption) error
	// RefreshJSON represents the tofu refresh subcommand with the `-json` flag.
	RefreshJSON(ctx context.Context, w io.Writer, opts ...RefreshOption) error
	// PlanJSON represents the tofu plan subcommand with the `-json` flag.
	PlanJSON(ctx context.Context, w io.Writer, opts ...PlanOption) error
	// Show reads the default state path and outputs the state as JSON.
	Show(ctx context.Context) (*tfjson.State, error)
	// ShowPlanFile reads a given plan file and outputs the plan as JSON.
	ShowPlanFile(ctx context.Context, planPath string) (*tfjson.Plan, error)
	// Output represents the tofu output subcommand.
	Output(ctx context.Context) (map[string]tfexec.OutputMeta, error)
	// Validate represents the tofu validate subcommand.
	Validate(ctx context.Context) (*tfjson.ValidateOutput, error)
	// StatePull represents the tofu state pull subcommand.
	StatePull(ctx context.Context) (string, error)
	// StateMv represents the tofu state mv subcommand.
	StateMv(ctx context.Context, source, destination string) error
	// StateRm represents the tofu state rm subcommand.
	StateRm(ctx context.Context, address string) error
}

// Command options mirror the tfexec option names the terraform workspace uses
// (Refresh, VarFile, Out, ...) so the tofu workspace builds the same option
// lists. Unlike tfexec's, they are small comparable structs, which lets the
// fake clients in tests assert on them with assert.Contains.

type (
	// ApplyOption configures the tofu apply subcommand.
	ApplyOption interface{ applyOption() }
	// PlanOption configures the tofu plan subcommand.
	PlanOption interface{ planOption() }
	// DestroyOption configures the tofu destroy subcommand.
	DestroyOption interface{ destroyOption() }
	// RefreshOption configures the tofu refresh subcommand.
	RefreshOption interface{ refreshOption() }
)

type optRefresh bool

func (optRefresh) applyOption()   {}
func (optRefresh) planOption()    {}
func (optRefresh) destroyOption() {}
func (optRefresh) refreshOption() {}

// Refresh mirrors tfexec.Refresh: passes -refresh=true to plan/apply/destroy.
func Refresh(b bool) optRefresh { return optRefresh(b) }

type optDestroy bool

func (optDestroy) applyOption() {}
func (optDestroy) planOption()  {}

// Destroy mirrors tfexec.Destroy: passes -destroy to apply (destroy via apply).
func Destroy(b bool) optDestroy { return optDestroy(b) }

type optRefreshOnly bool

func (optRefreshOnly) planOption() {}

// RefreshOnly mirrors tfexec.RefreshOnly: passes -refresh-only to plan.
func RefreshOnly(b bool) optRefreshOnly { return optRefreshOnly(b) }

type optTarget string

func (optTarget) planOption() {}

// Target mirrors tfexec.Target: passes -target=<addr> to plan.
func Target(t string) optTarget { return optTarget(t) }

type optVarFile string

func (optVarFile) applyOption()   {}
func (optVarFile) planOption()    {}
func (optVarFile) destroyOption() {}
func (optVarFile) refreshOption() {}

// VarFile mirrors tfexec.VarFile: passes -var-file=<path>.
func VarFile(fp string) optVarFile { return optVarFile(fp) }

type optOut string

func (optOut) planOption() {}

// Out mirrors tfexec.Out: passes -out=<path> to plan.
func Out(p string) optOut { return optOut(p) }

type optDirOrPlan string

func (optDirOrPlan) applyOption() {}

// DirOrPlan mirrors tfexec.DirOrPlan: passes the path to a saved plan file as
// the positional argument of apply.
func DirOrPlan(p string) optDirOrPlan { return optDirOrPlan(p) }

// applyOpts is the decoded form of the apply options.
type applyOpts struct {
	refresh  *bool
	destroy  *bool
	planFile string
	varFiles []string
}

func (o *applyOpts) add(opts ...ApplyOption) error {
	for _, opt := range opts {
		switch v := opt.(type) {
		case optRefresh:
			b := bool(v)
			o.refresh = &b
		case optDestroy:
			b := bool(v)
			o.destroy = &b
		case optDirOrPlan:
			o.planFile = string(v)
		case optVarFile:
			o.varFiles = append(o.varFiles, string(v))
		default:
			return fmt.Errorf("unrecognized apply option type %T", opt)
		}
	}
	return nil
}

func (o *applyOpts) args() ([]string, error) {
	if o.planFile != "" && o.destroy != nil && *o.destroy {
		return nil, errors.New("cannot apply a saved plan file with -destroy, use the plan's own destroy flag")
	}

	args := []string{"apply", "-auto-approve"}
	args = append(args, "-refresh="+boolArg(o.refresh))
	if o.destroy != nil && *o.destroy {
		args = append(args, "-destroy")
	}
	for _, fp := range o.varFiles {
		args = append(args, "-var-file="+fp)
	}
	args = append(args, "-json")
	if o.planFile != "" {
		// The plan file must be the final argument: OpenTofu rejects
		// flags that follow the positional plan file argument.
		args = append(args, o.planFile)
	}

	return args, nil
}

// planOpts is the decoded form of the plan options.
type planOpts struct {
	refresh     *bool
	destroy     *bool
	refreshOnly *bool
	out         string
	varFiles    []string
	targets     []string
}

func (o *planOpts) add(opts ...PlanOption) error {
	for _, opt := range opts {
		switch v := opt.(type) {
		case optRefresh:
			b := bool(v)
			o.refresh = &b
		case optDestroy:
			b := bool(v)
			o.destroy = &b
		case optRefreshOnly:
			b := bool(v)
			o.refreshOnly = &b
		case optOut:
			o.out = string(v)
		case optVarFile:
			o.varFiles = append(o.varFiles, string(v))
		case optTarget:
			o.targets = append(o.targets, string(v))
		default:
			return fmt.Errorf("unrecognized plan option type %T", opt)
		}
	}
	return nil
}

func (o *planOpts) args() ([]string, error) {
	args := []string{"plan"}
	args = append(args, "-refresh="+boolArg(o.refresh))
	if o.destroy != nil && *o.destroy {
		args = append(args, "-destroy")
	}
	if o.refreshOnly != nil && *o.refreshOnly {
		args = append(args, "-refresh-only")
	}
	if o.out != "" {
		args = append(args, "-out="+o.out)
	}
	for _, t := range o.targets {
		args = append(args, "-target="+t)
	}
	for _, fp := range o.varFiles {
		args = append(args, "-var-file="+fp)
	}
	args = append(args, "-json")

	return args, nil
}

// destroyOpts is the decoded form of the destroy options.
type destroyOpts struct {
	refresh  *bool
	varFiles []string
}

func (o *destroyOpts) add(opts ...DestroyOption) error {
	for _, opt := range opts {
		switch v := opt.(type) {
		case optRefresh:
			b := bool(v)
			o.refresh = &b
		case optVarFile:
			o.varFiles = append(o.varFiles, string(v))
		default:
			return fmt.Errorf("unrecognized destroy option type %T", opt)
		}
	}
	return nil
}

func (o *destroyOpts) args() ([]string, error) {
	args := []string{"destroy", "-auto-approve"}
	args = append(args, "-refresh="+boolArg(o.refresh))
	for _, fp := range o.varFiles {
		args = append(args, "-var-file="+fp)
	}
	args = append(args, "-json")

	return args, nil
}

// refreshOpts is the decoded form of the refresh options.
type refreshOpts struct {
	varFiles []string
}

func (o *refreshOpts) add(opts ...RefreshOption) error {
	for _, opt := range opts {
		switch v := opt.(type) {
		case optVarFile:
			o.varFiles = append(o.varFiles, string(v))
		default:
			return fmt.Errorf("unrecognized refresh option type %T", opt)
		}
	}
	return nil
}

func (o *refreshOpts) args() ([]string, error) {
	args := []string{"refresh"}
	for _, fp := range o.varFiles {
		args = append(args, "-var-file="+fp)
	}
	args = append(args, "-json")

	return args, nil
}

// boolArg renders an optional boolean flag, defaulting to terraform's own
// defaults when unset (refresh defaults to true everywhere).
func boolArg(b *bool) string {
	if b == nil {
		return "true"
	}
	if *b {
		return "true"
	}
	return "false"
}

// tofuClient executes the tofu CLI directly. tfexec cannot drive tofu because
// its version parser requires a `Terraform vX` banner while OpenTofu prints
// `OpenTofu vX`, so this client builds exec.Command calls itself.
type tofuClient struct {
	root     string
	execPath string

	// env is the exact environment handed to the child process, mirroring what
	// tfexec.SetEnv would have produced. See envSlice for the nil case.
	env map[string]string

	log hclog.Logger

	// version caches the parsed `tofu version` output.
	version *goversion.Version
}

func newTofuClient(root, execPath string, env map[string]string, log hclog.Logger) *tofuClient {
	if log == nil {
		log = hclog.NewNullLogger()
	}
	return &tofuClient{
		root:     root,
		execPath: execPath,
		env:      env,
		log:      log,
	}
}

// envSlice mirrors tfexec's SetEnv semantics: an explicitly set map becomes
// the child process environment verbatim (the workspace's LoadVariables
// already merged os.Environ into it), while nil falls back to os.Environ the
// same way tfexec's explicit-nil SetEnv does. It then applies the same
// hardening tfexec applies to every command: TF_IN_AUTOMATION=1, cleared
// TF_LOG* so inherited logging cannot pollute stderr, and TF_WORKSPACE
// removed so an inherited value cannot silently switch state workspaces.
func (t *tofuClient) envSlice() []string {
	env := make(map[string]string)
	if t.env == nil {
		for _, kv := range os.Environ() {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				continue
			}
			env[parts[0]] = parts[1]
		}
	} else {
		for k, v := range t.env {
			env[k] = v
		}
	}

	env["TF_IN_AUTOMATION"] = "1"
	env["TF_LOG"] = ""
	env["TF_LOG_CORE"] = ""
	env["TF_LOG_PATH"] = ""
	env["TF_LOG_PROVIDER"] = ""
	delete(env, "TF_WORKSPACE")

	slice := make([]string, 0, len(env))
	for k, v := range env {
		slice = append(slice, k+"="+v)
	}
	return slice
}

// run executes the tofu CLI with the given args, streaming stdout to the
// supplied writer and capturing stderr so failures carry the CLI's own
// diagnostics (the JSON stdout stream only carries diagnostics for commands
// that emit it).
func (t *tofuClient) run(ctx context.Context, stdout io.Writer, args ...string) error {
	t.log.Trace("executing tofu", "args", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, t.execPath, args...)
	cmd.Dir = t.root
	cmd.Env = t.envSlice()
	cmd.Stdout = stdout
	// Mirror tfexec: cancel sends SIGINT so tofu can clean up resources
	// applied mid-run; WaitDelay escalates to SIGKILL if it hangs.
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 60 * time.Second

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		if s := strings.TrimSpace(stderr.String()); s != "" {
			return fmt.Errorf("%w\n%s", err, s)
		}
		return err
	}

	return nil
}

// tofuVersionRE matches the first line of `tofu version` output:
// `OpenTofu vX.Y.Z` (optionally followed by more on the same line).
var tofuVersionRE = regexp.MustCompile(`(?m)^OpenTofu v([0-9]+\.[0-9]+\.[0-9]+\S*)`)

// parseTofuVersion parses `tofu version` output into a go-version. Unlike
// tfexec's version parser, which requires a `Terraform vX` banner and cannot
// parse OpenTofu's `OpenTofu vX` output, this accepts the OpenTofu banner.
func parseTofuVersion(out string) (*goversion.Version, error) {
	m := tofuVersionRE.FindStringSubmatch(out)
	if m == nil {
		return nil, fmt.Errorf("unable to parse tofu version from output: %s", strings.TrimSpace(out))
	}

	v, err := goversion.NewVersion(m[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse tofu version %q: %w", m[1], err)
	}
	return v, nil
}

func (t *tofuClient) Version(ctx context.Context, skipCache bool) (*goversion.Version, error) {
	if t.version != nil && !skipCache {
		return t.version, nil
	}

	var out bytes.Buffer
	if err := t.run(ctx, &out, "version"); err != nil {
		return nil, fmt.Errorf("unable to run tofu version: %w", err)
	}

	v, err := parseTofuVersion(out.String())
	if err != nil {
		return nil, err
	}
	t.version = v

	return v, nil
}

func (t *tofuClient) Init(ctx context.Context, backendConfig string) error {
	if err := t.run(ctx, io.Discard, "init", "-backend-config="+backendConfig); err != nil {
		return fmt.Errorf("unable to run tofu init: %w", err)
	}
	return nil
}

func (t *tofuClient) ApplyJSON(ctx context.Context, w io.Writer, opts ...ApplyOption) error {
	o := &applyOpts{}
	if err := o.add(opts...); err != nil {
		return err
	}

	args, err := o.args()
	if err != nil {
		return err
	}

	if err := t.run(ctx, w, args...); err != nil {
		return fmt.Errorf("unable to run tofu apply: %w", err)
	}
	return nil
}

func (t *tofuClient) DestroyJSON(ctx context.Context, w io.Writer, opts ...DestroyOption) error {
	o := &destroyOpts{}
	if err := o.add(opts...); err != nil {
		return err
	}

	args, err := o.args()
	if err != nil {
		return err
	}

	if err := t.run(ctx, w, args...); err != nil {
		return fmt.Errorf("unable to run tofu destroy: %w", err)
	}
	return nil
}

func (t *tofuClient) RefreshJSON(ctx context.Context, w io.Writer, opts ...RefreshOption) error {
	o := &refreshOpts{}
	if err := o.add(opts...); err != nil {
		return err
	}

	args, err := o.args()
	if err != nil {
		return err
	}

	if err := t.run(ctx, w, args...); err != nil {
		return fmt.Errorf("unable to run tofu refresh: %w", err)
	}
	return nil
}

func (t *tofuClient) PlanJSON(ctx context.Context, w io.Writer, opts ...PlanOption) error {
	o := &planOpts{}
	if err := o.add(opts...); err != nil {
		return err
	}

	args, err := o.args()
	if err != nil {
		return err
	}

	if err := t.run(ctx, w, args...); err != nil {
		return fmt.Errorf("unable to run tofu plan: %w", err)
	}
	return nil
}

func (t *tofuClient) Show(ctx context.Context) (*tfjson.State, error) {
	var out bytes.Buffer
	if err := t.run(ctx, &out, "show", "-json"); err != nil {
		return nil, fmt.Errorf("unable to run tofu show: %w", err)
	}

	state := &tfjson.State{}
	if err := unmarshalUseNumber(out.Bytes(), state); err != nil {
		return nil, fmt.Errorf("unable to parse tofu show output: %w", err)
	}
	return state, nil
}

func (t *tofuClient) ShowPlanFile(ctx context.Context, planPath string) (*tfjson.Plan, error) {
	var out bytes.Buffer
	if err := t.run(ctx, &out, "show", "-json", planPath); err != nil {
		return nil, fmt.Errorf("unable to run tofu show for plan file: %w", err)
	}

	plan := &tfjson.Plan{}
	if err := unmarshalUseNumber(out.Bytes(), plan); err != nil {
		return nil, fmt.Errorf("unable to parse tofu show output: %w", err)
	}
	return plan, nil
}

// unmarshalUseNumber decodes JSON with numbers as json.Number, matching
// tfexec: integers larger than 2^53 lose precision as float64.
func unmarshalUseNumber(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(v)
}

func (t *tofuClient) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	var out bytes.Buffer
	if err := t.run(ctx, &out, "output", "-json"); err != nil {
		return nil, fmt.Errorf("unable to run tofu output: %w", err)
	}

	outputs := map[string]tfexec.OutputMeta{}
	if err := json.Unmarshal(out.Bytes(), &outputs); err != nil {
		return nil, fmt.Errorf("unable to parse tofu output: %w", err)
	}
	return outputs, nil
}

func (t *tofuClient) Validate(ctx context.Context) (*tfjson.ValidateOutput, error) {
	var out bytes.Buffer
	if err := t.run(ctx, &out, "validate", "-json"); err != nil {
		return nil, fmt.Errorf("unable to run tofu validate: %w", err)
	}

	validateOutput := &tfjson.ValidateOutput{}
	if err := json.Unmarshal(out.Bytes(), validateOutput); err != nil {
		return nil, fmt.Errorf("unable to parse tofu validate output: %w", err)
	}
	return validateOutput, nil
}

func (t *tofuClient) StatePull(ctx context.Context) (string, error) {
	var out bytes.Buffer
	if err := t.run(ctx, &out, "state", "pull"); err != nil {
		return "", fmt.Errorf("unable to run tofu state pull: %w", err)
	}
	return out.String(), nil
}

func (t *tofuClient) StateMv(ctx context.Context, source, destination string) error {
	if err := t.run(ctx, io.Discard, "state", "mv", source, destination); err != nil {
		return fmt.Errorf("unable to run tofu state mv: %w", err)
	}
	return nil
}

func (t *tofuClient) StateRm(ctx context.Context, address string) error {
	if err := t.run(ctx, io.Discard, "state", "rm", address); err != nil {
		return fmt.Errorf("unable to run tofu state rm: %w", err)
	}
	return nil
}
