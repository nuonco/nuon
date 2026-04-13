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

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
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

// Workspace wraps the Pulumi Automation API for programmatic Pulumi operations.
type Workspace struct {
	stack   auto.Stack
	workDir string
	opts    *Options
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
	// Isolate Go build cache per workspace to prevent stale cache entries
	// from referencing cleaned-up temp directories of previous jobs.
	envVars["GOCACHE"] = filepath.Join(opts.WorkDir, ".go-cache")

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

	return &Workspace{
		stack:   stack,
		workDir: opts.WorkDir,
		opts:    opts,
	}, nil
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
func (w *Workspace) Preview(ctx context.Context) (*PreviewResult, error) {
	eventCh := make(chan events.EngineEvent, 100)

	// Drain events concurrently to prevent deadlock — Preview() writes
	// to the channel and blocks if the buffer fills before we read.
	var resourceChanges []ResourceChange
	var diagnostics []string
	done := make(chan struct{})

	go func() {
		defer close(done)
		for evt := range eventCh {
			if evt.ResourcePreEvent != nil {
				meta := evt.ResourcePreEvent.Metadata
				if meta.Type == "pulumi:pulumi:Stack" || strings.HasPrefix(meta.Type, "pulumi:providers:") {
					continue
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
	}()

	result, err := w.stack.Preview(ctx,
		optpreview.Message("Nuon preview"),
		optpreview.EventStreams(eventCh),
	)

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

// Up runs `pulumi up` and returns the result.
func (w *Workspace) Up(ctx context.Context) (*UpResult, error) {
	result, err := w.stack.Up(ctx,
		optup.Message("Nuon deploy"),
	)
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
	_, err := w.stack.Destroy(ctx)
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
