package installs

import (
	"context"
	"fmt"
	"os"
	"sort"
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

func (s *Service) Outputs(ctx context.Context, installID string, asJSON bool) error {
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

	// Stack outputs.
	if stack != nil && stack.InstallStackOutputs != nil {
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

	// Sandbox outputs — latest successful/active run.
	for _, run := range sandboxes {
		if (run.Status == "active" || run.Status == "succeeded") && run.Outputs != nil {
			out.Sandbox = run.Outputs
			break
		}
	}
	if out.Sandbox == nil && len(sandboxes) > 0 && sandboxes[0].Outputs != nil {
		out.Sandbox = sandboxes[0].Outputs
	}

	// Component outputs — fetch latest deploy for each in parallel.
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
				deploy, err := s.api.GetInstallComponentLatestDeploy(ctx, installID, comp.ComponentID)
				if err != nil || deploy == nil || deploy.Outputs == nil || len(deploy.Outputs) == 0 {
					return
				}
				cmu.Lock()
				out.Components[name] = deploy.Outputs
				cmu.Unlock()
			}(ic)
		}
		cwg.Wait()
	}

	if asJSON {
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
		view.Print("No outputs available for this install.")
	}

	if len(warnings) > 0 {
		fmt.Fprintln(os.Stderr)
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
	}

	return nil
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
