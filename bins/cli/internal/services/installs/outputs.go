package installs

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// OutputsOptions filters what `Outputs` returns. At most one of StackOnly,
// SandboxOnly, or ComponentID may be set — Cobra enforces this at the flag layer.
type OutputsOptions struct {
	StackOnly   bool
	SandboxOnly bool
	ComponentID string // id or name; empty means no component filter
}

type installOutputs struct {
	Stack      map[string]any            `json:"stack,omitempty"`
	Sandbox    map[string]any            `json:"sandbox,omitempty"`
	Components map[string]map[string]any `json:"components,omitempty"`
}

func (s *Service) Outputs(ctx context.Context, installID string, opts OutputsOptions, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	filtered := opts.StackOnly || opts.SandboxOnly || opts.ComponentID != ""
	wantStack := !filtered || opts.StackOnly
	wantSandbox := !filtered || opts.SandboxOnly
	wantComponents := !filtered || opts.ComponentID != ""

	var (
		stack      *models.AppInstallStack
		sandboxes  []*models.AppInstallSandboxRun
		components []*models.AppInstallComponent
		mu         sync.Mutex
		wg         sync.WaitGroup
		fetchErrs  []error
		warnings   []string
	)

	if wantStack {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stk, err := s.api.GetInstallStack(ctx, installID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("stack outputs unavailable: %s", err))
				return
			}
			stack = stk
		}()
	}
	if wantSandbox {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runs, _, err := s.api.GetInstallSandboxRuns(ctx, installID, &models.GetPaginatedQuery{Limit: 50})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("sandbox outputs unavailable: %s", err))
				return
			}
			sandboxes = runs
		}()
	}
	if wantComponents {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comps, _, err := s.api.GetInstallComponents(ctx, installID, &models.GetPaginatedQuery{Limit: 100})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fetchErrs = append(fetchErrs, err)
				return
			}
			components = comps
		}()
	}
	wg.Wait()

	if len(fetchErrs) > 0 {
		return ui.PrintError(fetchErrs[0])
	}

	out := installOutputs{}

	if wantStack && stack != nil && stack.InstallStackOutputs != nil {
		out.Stack = flattenStackOutputs(stack.InstallStackOutputs)
	}

	if wantSandbox {
		out.Sandbox = latestSandboxOutputs(sandboxes)
	}

	if wantComponents {
		out.Components = make(map[string]map[string]any)
		var cmu sync.Mutex
		var cwg sync.WaitGroup
		for _, ic := range components {
			ic := ic
			name := componentDisplayName(ic)

			// Skip if filtering and this isn't the target.
			if opts.ComponentID != "" &&
				!strings.EqualFold(name, opts.ComponentID) &&
				!strings.EqualFold(ic.ComponentID, opts.ComponentID) {
				continue
			}
			if ic.TerraformWorkspace == nil || ic.TerraformWorkspace.ID == "" {
				continue
			}

			cwg.Add(1)
			go func() {
				defer cwg.Done()
				outputs, err := s.getTerraformOutputs(ctx, ic.TerraformWorkspace.ID)
				cmu.Lock()
				defer cmu.Unlock()
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("component %s: error fetching outputs: %s", name, err))
					return
				}
				if len(outputs) == 0 {
					return
				}
				out.Components[name] = outputs
			}()
		}
		cwg.Wait()
	}

	if asJSON {
		switch {
		case opts.StackOnly:
			ui.PrintJSON(out.Stack)
		case opts.SandboxOnly:
			ui.PrintJSON(out.Sandbox)
		case opts.ComponentID != "":
			// Return the bare component map for the single match (if any).
			for _, v := range out.Components {
				ui.PrintJSON(v)
				return nil
			}
			ui.PrintJSON(map[string]any{})
		default:
			ui.PrintJSON(out)
		}
		return nil
	}

	header := lipgloss.NewStyle().Foreground(styles.AccentColor).Bold(true)
	view := ui.NewListView()
	empty := true
	printedSection := false

	if len(out.Stack) > 0 {
		empty = false
		printedSection = true
		fmt.Println(header.Render("Stack"))
		printOutputsTable(view, out.Stack)
	}

	if len(out.Sandbox) > 0 {
		empty = false
		if printedSection {
			fmt.Println()
		}
		printedSection = true
		fmt.Println(header.Render("Sandbox"))
		printOutputsTable(view, out.Sandbox)
	}

	compNames := make([]string, 0, len(out.Components))
	for name := range out.Components {
		compNames = append(compNames, name)
	}
	sort.Strings(compNames)
	for _, name := range compNames {
		empty = false
		if printedSection {
			fmt.Println()
		}
		printedSection = true
		fmt.Println(header.Render("Component " + name))
		printOutputsTable(view, out.Components[name])
	}

	if empty {
		switch {
		case opts.ComponentID != "":
			view.Print(fmt.Sprintf("No outputs found for component %q.", opts.ComponentID))
			availableComponents(components)
		case opts.SandboxOnly:
			view.Print("No sandbox outputs available for this install.")
		case opts.StackOnly:
			view.Print("No stack outputs available for this install.")
		default:
			view.Print("No outputs available for this install.")
		}
	}

	for _, w := range warnings {
		ui.PrintWarning(w)
	}

	return nil
}

func componentDisplayName(c *models.AppInstallComponent) string {
	if c.Component != nil && c.Component.Name != "" {
		return c.Component.Name
	}
	return c.ComponentID
}

func availableComponents(components []*models.AppInstallComponent) {
	if len(components) == 0 {
		return
	}
	names := make([]string, 0, len(components))
	for _, c := range components {
		names = append(names, componentDisplayName(c))
	}
	sort.Strings(names)
	fmt.Printf("\nAvailable components: %s\n", strings.Join(names, ", "))
}

// latestSandboxOutputs returns the outputs map from the most relevant sandbox
// run: the first active/succeeded run with outputs, otherwise the most recent
// run's outputs if any. Returns nil if no run has outputs.
func latestSandboxOutputs(runs []*models.AppInstallSandboxRun) map[string]any {
	for _, run := range runs {
		if (run.Status == "active" || run.Status == "succeeded") && run.Outputs != nil {
			return run.Outputs
		}
	}
	if len(runs) > 0 && runs[0].Outputs != nil {
		return runs[0].Outputs
	}
	return nil
}

func flattenStackOutputs(o *models.AppInstallStackOutputs) map[string]any {
	flat := make(map[string]any)
	if o.Aws != nil {
		aws := o.Aws
		if aws.AccountID != "" {
			flat["account_id"] = aws.AccountID
		}
		if aws.Region != "" {
			flat["region"] = aws.Region
		}
		if aws.VpcID != "" {
			flat["vpc_id"] = aws.VpcID
		}
		if aws.RunnerSubnet != "" {
			flat["runner_subnet"] = aws.RunnerSubnet
		}
		if len(aws.PrivateSubnets) > 0 {
			flat["private_subnets"] = aws.PrivateSubnets
		}
		if len(aws.PublicSubnets) > 0 {
			flat["public_subnets"] = aws.PublicSubnets
		}
		if aws.ProvisionIamRoleArn != "" {
			flat["provision_iam_role_arn"] = aws.ProvisionIamRoleArn
		}
		if aws.DeprovisionIamRoleArn != "" {
			flat["deprovision_iam_role_arn"] = aws.DeprovisionIamRoleArn
		}
		if aws.MaintenanceIamRoleArn != "" {
			flat["maintenance_iam_role_arn"] = aws.MaintenanceIamRoleArn
		}
		if aws.RunnerIamRoleArn != "" {
			flat["runner_iam_role_arn"] = aws.RunnerIamRoleArn
		}
		if len(aws.BreakGlassRoleArns) > 0 {
			flat["break_glass_role_arns"] = aws.BreakGlassRoleArns
		}
		if len(aws.CustomRoleArns) > 0 {
			flat["custom_role_arns"] = aws.CustomRoleArns
		}
	}
	for k, v := range o.Data {
		flat[k] = v
	}
	for k, v := range o.DataContents {
		flat[k] = v
	}
	return flat
}

// getTerraformOutputs fetches the latest terraform state for a workspace and
// extracts the output values. It tries the state-json endpoint first (already
// in `terraform show -json` shape) and falls back to the raw state.
func (s *Service) getTerraformOutputs(ctx context.Context, workspaceID string) (map[string]any, error) {
	raw, err := s.api.GetTerraformWorkspaceLatestStateJSON(ctx, workspaceID)
	if err == nil && len(raw) > 0 {
		if outputs := parseTerraformShowOutputs(raw); len(outputs) > 0 {
			return outputs, nil
		}
		if outputs, parseErr := parseRawTerraformStateOutputs(raw); parseErr == nil && len(outputs) > 0 {
			return outputs, nil
		}
	}

	state, stateErr := s.api.GetTerraformWorkspaceLatestState(ctx, workspaceID)
	if stateErr == nil && state != nil && len(state.Contents) > 0 {
		raw := int64SliceToBytes(state.Contents)
		if outputs := parseTerraformShowOutputs(raw); len(outputs) > 0 {
			return outputs, nil
		}
		return parseRawTerraformStateOutputs(raw)
	}

	if err != nil {
		return nil, fmt.Errorf("state-json: %w", err)
	}
	if stateErr != nil {
		return nil, fmt.Errorf("raw state: %w", stateErr)
	}
	return nil, fmt.Errorf("no state data available")
}

func int64SliceToBytes(s []int64) []byte {
	b := make([]byte, len(s))
	for i, v := range s {
		b[i] = byte(v)
	}
	return b
}

// parseTerraformShowOutputs parses `terraform show -json` format:
// {values: {outputs: {name: {value, type}}}}.
func parseTerraformShowOutputs(raw []byte) map[string]any {
	var tfShow struct {
		Values struct {
			Outputs map[string]struct {
				Value any `json:"value"`
				Type  any `json:"type"`
			} `json:"outputs"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &tfShow); err != nil {
		return nil
	}
	result := make(map[string]any, len(tfShow.Values.Outputs))
	for k, v := range tfShow.Values.Outputs {
		result[k] = v.Value
	}
	return result
}

// parseRawTerraformStateOutputs parses raw terraform state:
// {outputs: {name: {value, type}}}.
func parseRawTerraformStateOutputs(raw []byte) (map[string]any, error) {
	var tfState struct {
		Outputs map[string]struct {
			Value any    `json:"value"`
			Type  string `json:"type"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(raw, &tfState); err != nil {
		return nil, fmt.Errorf("parsing terraform state: %w", err)
	}
	if len(tfState.Outputs) == 0 {
		return nil, nil
	}
	result := make(map[string]any, len(tfState.Outputs))
	for k, v := range tfState.Outputs {
		result[k] = v.Value
	}
	return result, nil
}

func printOutputsTable(view *ui.ListView, m map[string]any) {
	flat := make(map[string]string)
	flattenMap("", m, flat)

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := [][]string{{"KEY", "VALUE"}}
	for _, k := range keys {
		data = append(data, []string{k, flat[k]})
	}
	view.Render(data)
}

// flattenMap recursively flattens a nested map into dot-notation keys.
func flattenMap(prefix string, m map[string]any, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flattenMap(key, val, out)
		case []any:
			parts := make([]string, len(val))
			for i, item := range val {
				parts[i] = fmt.Sprintf("%v", item)
			}
			out[key] = strings.Join(parts, ", ")
		default:
			out[key] = fmt.Sprintf("%v", v)
		}
	}
}
