package installs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type installOutputs struct {
	Stack      map[string]any            `json:"stack,omitempty"`
	Sandbox    map[string]any            `json:"sandbox,omitempty"`
	Components map[string]map[string]any `json:"components,omitempty"`
}

func (s *Service) Outputs(ctx context.Context, installID, componentFilter string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// Fetch all data in parallel.
	var (
		stack      *models.AppInstallStack
		sandboxes  []*models.AppInstallSandboxRun
		components []*models.AppInstallComponent
		mu         sync.Mutex
		wg         sync.WaitGroup
		fetchErrs  []error
	)

	var warnings []string
	wg.Add(3)
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
	wg.Wait()

	if len(fetchErrs) > 0 {
		return ui.PrintError(fetchErrs[0])
	}

	out := installOutputs{
		Components: make(map[string]map[string]any),
	}

	// Stack outputs (skip when filtering by component).
	if componentFilter == "" && stack != nil && stack.InstallStackOutputs != nil {
		outputs := stack.InstallStackOutputs
		flat := make(map[string]any)
		if outputs.Aws != nil {
			aws := outputs.Aws
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
		if outputs.Data != nil {
			for k, v := range outputs.Data {
				flat[k] = v
			}
		}
		if outputs.DataContents != nil {
			for k, v := range outputs.DataContents {
				flat[k] = v
			}
		}
		out.Stack = flat
	}

	// Sandbox outputs (skip when filtering by component).
	if componentFilter == "" {
		for _, run := range sandboxes {
			if (run.Status == "active" || run.Status == "succeeded") && run.Outputs != nil {
				out.Sandbox = run.Outputs
				break
			}
		}
		if out.Sandbox == nil && len(sandboxes) > 0 && sandboxes[0].Outputs != nil {
			out.Sandbox = sandboxes[0].Outputs
		}
	}

	// Component outputs — fetch terraform state outputs via workspace state JSON.
	if len(components) > 0 {
		var cmu sync.Mutex
		var cwg sync.WaitGroup
		cwg.Add(len(components))
		for _, ic := range components {
			go func(comp *models.AppInstallComponent) {
				defer cwg.Done()
				name := comp.ComponentID
				if comp.Component != nil && comp.Component.Name != "" {
					name = comp.Component.Name
				}

				// Skip if filtering and this isn't the target.
				if componentFilter != "" &&
					!strings.EqualFold(name, componentFilter) &&
					!strings.EqualFold(comp.ComponentID, componentFilter) {
					return
				}

				if comp.TerraformWorkspace == nil || comp.TerraformWorkspace.ID == "" {
					return
				}

				outputs, err := s.getTerraformOutputs(ctx, comp.TerraformWorkspace.ID)
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
			}(ic)
		}
		cwg.Wait()
	}

	if asJSON {
		if componentFilter != "" && len(out.Components) == 1 {
			// When filtering by component in JSON mode, return just that component's outputs.
			for _, v := range out.Components {
				ui.PrintJSON(v)
				return nil
			}
		}
		ui.PrintJSON(out)
		return nil
	}

	// Human-friendly table output grouped by section.
	view := ui.NewListView()
	empty := true

	if len(out.Stack) > 0 {
		empty = false
		fmt.Println("Stack:")
		flat := make(map[string]string)
		flattenMap("", out.Stack, flat)
		printSection(view, flat)
	}

	if len(out.Sandbox) > 0 {
		empty = false
		if len(out.Stack) > 0 {
			fmt.Println()
		}
		fmt.Println("Sandbox:")
		flat := make(map[string]string)
		flattenMap("", out.Sandbox, flat)
		printSection(view, flat)
	}

	compNames := make([]string, 0, len(out.Components))
	for name := range out.Components {
		compNames = append(compNames, name)
	}
	sort.Strings(compNames)
	for _, name := range compNames {
		empty = false
		fmt.Printf("\nComponent %s:\n", name)
		flat := make(map[string]string)
		flattenMap("", out.Components[name], flat)
		printSection(view, flat)
	}

	if empty {
		if componentFilter != "" {
			fmt.Printf("No outputs found for component %q.\n", componentFilter)
			// List available components.
			names := make([]string, 0, len(components))
			for _, c := range components {
				name := c.ComponentID
				if c.Component != nil && c.Component.Name != "" {
					name = c.Component.Name
				}
				names = append(names, name)
			}
			if len(names) > 0 {
				sort.Strings(names)
				fmt.Printf("\nAvailable components: %s\n", strings.Join(names, ", "))
			}
		} else {
			view.Print("No outputs available for this install.")
		}
	}

	if len(warnings) > 0 {
		fmt.Fprintln(os.Stderr)
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
	}

	return nil
}

// getTerraformOutputs fetches the latest terraform state JSON for a workspace
// and extracts the output values.
func (s *Service) getTerraformOutputs(ctx context.Context, workspaceID string) (map[string]any, error) {
	// Try state-json by ID — returns terraform show -json directly.
	raw, err := s.api.GetTerraformWorkspaceLatestStateJSON(ctx, workspaceID)
	if err == nil && len(raw) > 0 {
		if outputs := parseTerraformShowOutputs(raw); len(outputs) > 0 {
			return outputs, nil
		}
		if outputs, parseErr := parseRawTerraformStateOutputs(raw); parseErr == nil && len(outputs) > 0 {
			return outputs, nil
		}
	}

	// Fall back to raw state by ID.
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

// parseTerraformShowOutputs parses `terraform show -json` format: {values: {outputs: {name: {value, type}}}}
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

// parseRawTerraformStateOutputs parses raw terraform state: {outputs: {name: {value, type}}}
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

func printSection(view *ui.ListView, flat map[string]string) {
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
