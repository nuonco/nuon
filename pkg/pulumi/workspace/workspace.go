package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// StateBackend configures the state backend for Pulumi.
// Uses the Nuon control plane HTTP backend (reuses TerraformWorkspace infrastructure).
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
	Config       map[string]string
	EnvVars      map[string]string
	StateBackend *StateBackend
}

// PreviewResult contains the output of a pulumi preview.
type PreviewResult struct {
	StdOut        string         `json:"stdout"`
	StdErr        string         `json:"stderr"`
	ChangeSummary map[string]int `json:"change_summary"`
}

// UpResult contains the output of a pulumi up.
type UpResult struct {
	StdOut  string                 `json:"stdout"`
	StdErr  string                 `json:"stderr"`
	Outputs map[string]interface{} `json:"outputs"`
}

// Workspace wraps the Pulumi Automation API for programmatic Pulumi operations.
type Workspace struct {
	stack   auto.Stack
	workDir string
	opts    *Options
}

// New creates a new Pulumi workspace with a local file backend for state.
func New(ctx context.Context, opts *Options) (*Workspace, error) {
	// Set up a local state backend directory within the work directory.
	// State will be synced to/from the control plane before/after operations.
	stateDir := filepath.Join(opts.WorkDir, ".pulumi-state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create state directory: %w", err)
	}

	// Set PULUMI_BACKEND_URL to use local file backend
	envVars := make(map[string]string)
	for k, v := range opts.EnvVars {
		envVars[k] = v
	}
	envVars["PULUMI_BACKEND_URL"] = fmt.Sprintf("file://%s", stateDir)
	// Disable automatic update checks
	envVars["PULUMI_SKIP_UPDATE_CHECK"] = "true"

	// Detect the project name from Pulumi.yaml if present
	projectName := "nuon-project"
	pulumiYamlPath := filepath.Join(opts.WorkDir, "Pulumi.yaml")
	if data, err := os.ReadFile(pulumiYamlPath); err == nil {
		var proj workspace.Project
		if err := json.Unmarshal(data, &proj); err == nil && proj.Name != "" {
			projectName = string(proj.Name)
		}
	}

	// Create or select the stack using the local workspace
	stack, err := auto.UpsertStackLocalSource(ctx, opts.StackName, opts.WorkDir,
		auto.Project(workspace.Project{
			Name:    tokens.PackageName(projectName),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		}),
		auto.EnvVars(envVars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create/select stack: %w", err)
	}

	// Set stack config values
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

// Preview runs `pulumi preview` and returns the result.
func (w *Workspace) Preview(ctx context.Context) (*PreviewResult, error) {
	result, err := w.stack.Preview(ctx,
		optpreview.Message("Nuon preview"),
	)
	if err != nil {
		return nil, fmt.Errorf("pulumi preview failed: %w", err)
	}

	changeSummary := make(map[string]int)
	for k, v := range result.ChangeSummary {
		changeSummary[string(k)] = v
	}

	return &PreviewResult{
		StdOut:        result.StdOut,
		StdErr:        result.StdErr,
		ChangeSummary: changeSummary,
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

	outputs := make(map[string]interface{})
	for k, v := range result.Outputs {
		outputs[k] = v.Value
	}

	return &UpResult{
		StdOut:  result.StdOut,
		StdErr:  result.StdErr,
		Outputs: outputs,
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
func (w *Workspace) Outputs(ctx context.Context) (map[string]interface{}, error) {
	outs, err := w.stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get stack outputs: %w", err)
	}

	result := make(map[string]interface{})
	for k, v := range outs {
		result[k] = v.Value
	}
	return result, nil
}
