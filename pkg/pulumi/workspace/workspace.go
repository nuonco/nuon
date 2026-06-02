package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"go.uber.org/zap"
)

// StateBackend configures the state backend for Pulumi.
type StateBackend struct {
	APIEndpoint string
	WorkspaceID string
	Token       string
	JobID       string
}

// Options configures a Pulumi workspace.
type Options struct {
	WorkDir      string
	StackName    string
	Runtime      string
	Config       map[string]string
	EnvVars      map[string]string
	StateBackend *StateBackend
	Logger       *zap.Logger
}

// ResourceChange represents a single resource change in a preview, structured
// to mirror terraform's resource_changes for UI parity.
type ResourceChange struct {
	URN          string                          `json:"urn"`
	Type         string                          `json:"type"`
	Name         string                          `json:"name"`
	Action       string                          `json:"action"`
	Diffs        []string                        `json:"diffs,omitempty"`
	DetailedDiff map[string]apitype.PropertyDiff `json:"detailed_diff,omitempty"`
	OldInputs    map[string]any                  `json:"old_inputs,omitempty"`
	NewInputs    map[string]any                  `json:"new_inputs,omitempty"`
	Provider     string                          `json:"provider,omitempty"`
}

// PreviewResult contains the output of a pulumi preview with structured
// resource changes comparable to terraform's plan JSON.
type PreviewResult struct {
	StdOut          string           `json:"stdout"`
	StdErr          string           `json:"stderr"`
	ChangeSummary   map[string]int   `json:"change_summary"`
	ResourceChanges []ResourceChange `json:"resource_changes,omitempty"`
	Diagnostics     []string         `json:"diagnostics,omitempty"`
}

// UpResult contains the output of a pulumi up.
type UpResult struct {
	StdOut  string         `json:"stdout"`
	StdErr  string         `json:"stderr"`
	Outputs map[string]any `json:"outputs"`
}

// PreviewOpts are optional knobs for Preview.
type PreviewOpts struct {
	// PlanOutPath, if set, makes Pulumi save an update plan to this path so
	// a subsequent Up can replay it deterministically.
	PlanOutPath string
}

// UpOpts are optional knobs for Up.
type UpOpts struct {
	// PlanInPath, if set, makes Pulumi apply this previously-saved update
	// plan instead of computing a new diff. Up will fail if reality has
	// drifted from what the plan expected.
	PlanInPath string
}

// Workspace wraps the Pulumi Automation API for programmatic Pulumi operations.
type Workspace struct {
	stack   auto.Stack
	workDir string
	opts    *Options
	logger  *zap.Logger
}

// New creates a new Pulumi workspace with a local file backend for state.
func New(ctx context.Context, opts *Options) (*Workspace, error) {
	stateDir := filepath.Join(opts.WorkDir, ".pulumi-state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create state directory: %w", err)
	}

	envVars := make(map[string]string)
	for k, v := range opts.EnvVars {
		envVars[k] = v
	}
	envVars["PULUMI_BACKEND_URL"] = fmt.Sprintf("file://%s", stateDir)
	hash := sha256.Sum256([]byte("nuon-pulumi:" + opts.StackName))
	envVars["PULUMI_CONFIG_PASSPHRASE"] = hex.EncodeToString(hash[:])
	envVars["PULUMI_SKIP_UPDATE_CHECK"] = "true"
	// Required by the Pulumi CLI to accept --save-plan / --plan in this version.
	envVars["PULUMI_EXPERIMENTAL"] = "true"
	// Fixed shared path so every job reuses the compiled pulumi-gcp SDK. Deriving
	// from WorkDir broke this for component deploys (their WorkDir is a per-job
	// temp dir), forcing a full recompile on every preview and apply.
	goBuildCache := filepath.Join(os.TempDir(), "nuon-pulumi-go-cache")
	if err := os.MkdirAll(goBuildCache, 0755); err != nil {
		return nil, fmt.Errorf("unable to create go build cache dir: %w", err)
	}
	if _, ok := envVars["GOCACHE"]; !ok {
		envVars["GOCACHE"] = goBuildCache
	}

	projectName := "nuon-project"
	pulumiYamlPath := filepath.Join(opts.WorkDir, "Pulumi.yaml")
	if data, err := os.ReadFile(pulumiYamlPath); err == nil {
		var proj workspace.Project
		if err := json.Unmarshal(data, &proj); err == nil && proj.Name != "" {
			projectName = string(proj.Name)
		}
	}

	runtime := opts.Runtime
	if runtime == "" {
		runtime = "go"
	}

	stack, err := auto.UpsertStackLocalSource(ctx, opts.StackName, opts.WorkDir,
		auto.Project(workspace.Project{
			Name:    tokens.PackageName(projectName),
			Runtime: workspace.NewProjectRuntimeInfo(runtime, nil),
		}),
		auto.EnvVars(envVars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create/select stack: %w", err)
	}

	for k, v := range opts.Config {
		if err := stack.SetConfig(ctx, k, auto.ConfigValue{Value: v}); err != nil {
			return nil, fmt.Errorf("unable to set config %s: %w", k, err)
		}
	}

	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Workspace{
		stack:   stack,
		workDir: opts.WorkDir,
		opts:    opts,
		logger:  logger,
	}, nil
}

// progressHeartbeat is how often in-flight resources re-log a "still <op>"
// line so long-running creates (e.g. a GKE cluster) show liveness.
const progressHeartbeat = 10 * time.Second

type inflightResource struct {
	op      string
	name    string
	typ     string
	started time.Time
}

// logProgressEvents drains engine events, logging a line per user resource as
// it starts and finishes, plus a periodic "still <op>" heartbeat for resources
// that are still in flight — so plan/apply progress stays visible in the job
// log stream even during long single-resource waits.
func (w *Workspace) logProgressEvents(eventCh <-chan events.EngineEvent, done chan<- struct{}, collect func(events.EngineEvent)) {
	defer close(done)

	var mu sync.Mutex
	inflight := map[string]inflightResource{}

	ticker := time.NewTicker(progressHeartbeat)
	defer ticker.Stop()
	stopTicker := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopTicker:
				return
			case <-ticker.C:
				mu.Lock()
				for _, r := range inflight {
					w.logger.Info(fmt.Sprintf("still %s %s (%s) [%s]", opGerund(r.op), r.name, r.typ, time.Since(r.started).Round(time.Second)))
				}
				mu.Unlock()
			}
		}
	}()

	for evt := range eventCh {
		if collect != nil {
			collect(evt)
		}
		switch {
		case evt.ResourcePreEvent != nil:
			m := evt.ResourcePreEvent.Metadata
			if skipResourceLog(m.Type, string(m.Op)) {
				continue
			}
			mu.Lock()
			inflight[m.URN] = inflightResource{op: string(m.Op), name: extractResourceName(m.URN), typ: m.Type, started: time.Now()}
			mu.Unlock()
			w.logger.Info(fmt.Sprintf("%s %s (%s)", opGerund(string(m.Op)), extractResourceName(m.URN), m.Type))
		case evt.ResOutputsEvent != nil:
			m := evt.ResOutputsEvent.Metadata
			if skipResourceLog(m.Type, string(m.Op)) {
				continue
			}
			mu.Lock()
			delete(inflight, m.URN)
			mu.Unlock()
			w.logger.Info(fmt.Sprintf("%s %s (%s)", opPast(string(m.Op)), extractResourceName(m.URN), m.Type))
		}
	}
	close(stopTicker)
}

func skipResourceLog(typ, op string) bool {
	if typ == "pulumi:pulumi:Stack" || strings.HasPrefix(typ, "pulumi:providers:") {
		return true
	}
	// "same" means no change — don't spam the log with unchanged resources.
	return op == "same"
}

func opGerund(op string) string {
	switch op {
	case "create", "create-replacement":
		return "creating"
	case "update":
		return "updating"
	case "delete", "delete-replaced":
		return "deleting"
	case "replace":
		return "replacing"
	case "refresh":
		return "refreshing"
	case "import", "import-replacement":
		return "importing"
	case "read":
		return "reading"
	default:
		return op
	}
}

func opPast(op string) string {
	switch op {
	case "create", "create-replacement":
		return "created"
	case "update":
		return "updated"
	case "delete", "delete-replaced":
		return "deleted"
	case "replace":
		return "replaced"
	case "refresh":
		return "refreshed"
	case "import", "import-replacement":
		return "imported"
	case "read":
		return "read"
	default:
		return op + " done"
	}
}

// StateDir returns the path to the local state backend directory.
func (w *Workspace) StateDir() string {
	return filepath.Join(w.workDir, ".pulumi-state")
}

// extractResourceName extracts the resource name from a Pulumi URN.
// URN format: urn:pulumi:stack::project::type::name
func extractResourceName(urn string) string {
	parts := strings.Split(urn, "::")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return urn
}

// Preview runs `pulumi preview` with event streaming to capture structured
// per-resource changes, producing output comparable to terraform plan JSON.
// When opts.PlanOutPath is set, a Pulumi update plan is also written to that
// path; pass it back via UpOpts.PlanInPath on a subsequent Up to skip the
// implicit re-preview and guard against drift between plan and apply.
func (w *Workspace) Preview(ctx context.Context, opts *PreviewOpts) (*PreviewResult, error) {
	eventCh := make(chan events.EngineEvent, 100)

	// Drain events concurrently to prevent deadlock — Preview() writes
	// to the channel and blocks if the buffer fills before we read.
	var resourceChanges []ResourceChange
	var diagnostics []string
	done := make(chan struct{})

	collect := func(evt events.EngineEvent) {
		if evt.ResourcePreEvent != nil {
			meta := evt.ResourcePreEvent.Metadata
			if meta.Type == "pulumi:pulumi:Stack" || strings.HasPrefix(meta.Type, "pulumi:providers:") {
				return
			}
			rc := ResourceChange{
				URN:          meta.URN,
				Type:         meta.Type,
				Name:         extractResourceName(meta.URN),
				Action:       string(meta.Op),
				Diffs:        meta.Diffs,
				DetailedDiff: meta.DetailedDiff,
				Provider:     meta.Provider,
			}
			if meta.Old != nil {
				rc.OldInputs = meta.Old.Inputs
			}
			if meta.New != nil {
				rc.NewInputs = meta.New.Inputs
			}
			resourceChanges = append(resourceChanges, rc)
		}
		if evt.DiagnosticEvent != nil {
			sev := evt.DiagnosticEvent.Severity
			if sev == "warning" || sev == "error" {
				msg := strings.TrimSpace(evt.DiagnosticEvent.Message)
				if msg != "" {
					diagnostics = append(diagnostics, msg)
				}
			}
		}
	}

	go w.logProgressEvents(eventCh, done, collect)

	previewOpts := []optpreview.Option{
		optpreview.Message("Nuon preview"),
		optpreview.EventStreams(eventCh),
	}
	if opts != nil && opts.PlanOutPath != "" {
		previewOpts = append(previewOpts, optpreview.Plan(opts.PlanOutPath))
	}
	result, err := w.stack.Preview(ctx, previewOpts...)

	// Wait for all events to be processed
	<-done

	if err != nil {
		return nil, fmt.Errorf("pulumi preview failed: %w", err)
	}

	// Build change summary from our filtered resource changes (excludes
	// Stack and provider resources) instead of Pulumi's raw summary.
	changeSummary := make(map[string]int)
	for _, rc := range resourceChanges {
		changeSummary[rc.Action]++
	}

	return &PreviewResult{
		StdOut:          result.StdOut,
		StdErr:          result.StdErr,
		ChangeSummary:   changeSummary,
		ResourceChanges: resourceChanges,
		Diagnostics:     diagnostics,
	}, nil
}

// Up runs `pulumi up` and returns the result. When opts.PlanInPath is set,
// Pulumi applies that previously-saved update plan instead of computing a
// fresh diff — faster, and refuses to apply if reality has drifted.
func (w *Workspace) Up(ctx context.Context, opts *UpOpts) (*UpResult, error) {
	eventCh := make(chan events.EngineEvent, 100)
	done := make(chan struct{})
	go w.logProgressEvents(eventCh, done, nil)

	upOpts := []optup.Option{
		optup.Message("Nuon deploy"),
		optup.EventStreams(eventCh),
	}
	if opts != nil && opts.PlanInPath != "" {
		upOpts = append(upOpts, optup.Plan(opts.PlanInPath))
	} else {
		// No saved plan: refresh against reality first so already-existing or
		// drifted resources are adopted/reconciled instead of recreated.
		upOpts = append(upOpts, optup.Refresh())
	}
	result, err := w.stack.Up(ctx, upOpts...)
	<-done
	if err != nil {
		return nil, fmt.Errorf("pulumi up failed: %w", err)
	}

	outputs := make(map[string]any)
	for k, v := range result.Outputs {
		outputs[k] = v.Value
	}

	return &UpResult{
		StdOut:  result.StdOut,
		StdErr:  result.StdErr,
		Outputs: outputs,
	}, nil
}

// DestroyPreview generates a synthetic preview of what destroy would do
// by inspecting the current stack state.
func (w *Workspace) DestroyPreview(ctx context.Context) (*PreviewResult, error) {
	stateJSON, err := w.ExportState(ctx)
	if err != nil {
		return &PreviewResult{
			StdOut:        "No resources to destroy (no state found)",
			ChangeSummary: map[string]int{"delete": 0},
		}, nil
	}

	var state struct {
		Resources []struct {
			URN    string         `json:"urn"`
			Type   string         `json:"type"`
			Inputs map[string]any `json:"inputs"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return &PreviewResult{
			StdOut:        "Unable to parse state for destroy preview",
			StdErr:        err.Error(),
			ChangeSummary: map[string]int{},
		}, nil
	}

	var resourceChanges []ResourceChange
	var lines []string
	for _, r := range state.Resources {
		if r.Type == "pulumi:pulumi:Stack" || strings.HasPrefix(r.Type, "pulumi:providers:") {
			continue
		}
		resourceChanges = append(resourceChanges, ResourceChange{
			URN:       r.URN,
			Type:      r.Type,
			Name:      extractResourceName(r.URN),
			Action:    "delete",
			OldInputs: r.Inputs,
		})
		lines = append(lines, fmt.Sprintf(" -  %s delete", r.URN))
	}

	deleteCount := len(resourceChanges)
	stdout := fmt.Sprintf("Previewing destroy (resources to be deleted: %d):\n\n%s\n\nResources:\n    - %d to delete\n",
		deleteCount, strings.Join(lines, "\n"), deleteCount)

	return &PreviewResult{
		StdOut:          stdout,
		ChangeSummary:   map[string]int{"delete": deleteCount},
		ResourceChanges: resourceChanges,
	}, nil
}

// Destroy runs `pulumi destroy`.
func (w *Workspace) Destroy(ctx context.Context) error {
	eventCh := make(chan events.EngineEvent, 100)
	done := make(chan struct{})
	go w.logProgressEvents(eventCh, done, nil)

	_, err := w.stack.Destroy(ctx, optdestroy.EventStreams(eventCh))
	<-done
	if err != nil {
		return fmt.Errorf("pulumi destroy failed: %w", err)
	}
	return nil
}

// ExportState exports the current stack state as JSON bytes.
func (w *Workspace) ExportState(ctx context.Context) ([]byte, error) {
	deployment, err := w.stack.Export(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to export stack state: %w", err)
	}
	return deployment.Deployment, nil
}

// ImportState imports stack state from JSON bytes.
func (w *Workspace) ImportState(ctx context.Context, stateJSON []byte) error {
	deployment := apitype.UntypedDeployment{
		Version:    3,
		Deployment: stateJSON,
	}
	if err := w.stack.Import(ctx, deployment); err != nil {
		return fmt.Errorf("unable to import stack state: %w", err)
	}
	return nil
}

// EncryptionSalt returns the stack's per-stack encryption salt. Pulumi
// generates this lazily the first time it encrypts a value with the
// passphrase secrets manager, and persists it in the stack config file.
// Returns "" if no salt has been generated yet.
func (w *Workspace) EncryptionSalt(ctx context.Context) (string, error) {
	settings, err := w.stack.Workspace().StackSettings(ctx, w.opts.StackName)
	if err != nil {
		return "", fmt.Errorf("unable to read stack settings: %w", err)
	}
	if settings == nil {
		return "", nil
	}
	return settings.EncryptionSalt, nil
}

// SetEncryptionSalt writes salt onto the stack's config file so that secret
// values encrypted under that salt elsewhere (e.g. an update plan saved by a
// previous job) can be decrypted here. No-op when salt is empty.
func (w *Workspace) SetEncryptionSalt(ctx context.Context, salt string) error {
	if salt == "" {
		return nil
	}
	settings, err := w.stack.Workspace().StackSettings(ctx, w.opts.StackName)
	if err != nil {
		return fmt.Errorf("unable to read stack settings: %w", err)
	}
	if settings == nil {
		settings = &workspace.ProjectStack{}
	}
	if settings.EncryptionSalt == salt {
		return nil
	}
	settings.EncryptionSalt = salt
	if err := w.stack.Workspace().SaveStackSettings(ctx, w.opts.StackName, settings); err != nil {
		return fmt.Errorf("unable to save stack settings: %w", err)
	}
	return nil
}

// Outputs returns the current stack outputs.
func (w *Workspace) Outputs(ctx context.Context) (map[string]any, error) {
	outs, err := w.stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get stack outputs: %w", err)
	}

	result := make(map[string]any)
	for k, v := range outs {
		result[k] = v.Value
	}
	return result, nil
}
