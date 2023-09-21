package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) getCurrentOrg(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		Install: true,
		Json:    true,
		Args: []string{
			"-j",
			"orgs",
			"current",
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute get org: %w", err)
	}

	return nil
}

func (w *wkflow) getApp(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"apps",
			"get",
			"-a",
			outputs.AppID,
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute get app: %w", err)
	}

	return nil
}

func (w *wkflow) listApps(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"apps",
			"list",
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute list apps: %w", err)
	}

	return nil
}

func (w *wkflow) listComponentBuilds(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    true,
		Args: []string{
			"builds",
			"list",
			"-c",
			outputs.ComponentIDs[0],
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to list component builds: %w", err)
	}

	return nil
}

func (w *wkflow) buildComponent(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    false,
		Args: []string{
			"builds",
			"create",
			"-c",
			outputs.ComponentIDs[0],
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to build component: %w", err)
	}

	return nil
}

func (w *wkflow) getComponent(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"components",
			"get",
			"-c",
			outputs.ComponentIDs[0],
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute get component: %w", err)
	}

	return nil
}

func (w *wkflow) listComponents(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"components",
			"list",
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute list components: %w", err)
	}
	w.l.Info("list components", zap.Any("response", resp))

	return nil
}

func (w *wkflow) getInstall(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	if len(outputs.InstallIDs) < 1 {
		return fmt.Errorf("invalid e2e terraform, at least one install id must be exported")
	}

	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"installs",
			"get",
			"-i",
			outputs.InstallIDs[0],
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute list components: %w", err)
	}
	w.l.Info("list components", zap.Any("response", resp))

	return nil
}

func (w *wkflow) listInstalls(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	var resp activities.CLICommandResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
		OrgID:   outputs.OrgID,
		AppID:   outputs.AppID,
		Install: false,
		Json:    true,
		Args: []string{
			"-j",
			"installs",
			"list",
		},
	}, &resp); err != nil {
		return fmt.Errorf("unable to execute list components: %w", err)
	}
	w.l.Info("list components", zap.Any("response", resp))

	return nil
}

func (w *wkflow) buildAndReleaseComponents(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	for _, compID := range outputs.ComponentIDs {
		var buildResp activities.CLICommandResponse
		if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
			OrgID:   outputs.OrgID,
			AppID:   outputs.AppID,
			Install: false,
			Json:    true,
			Args: []string{
				"-j",
				"builds",
				"create",
				"-c",
				compID,
			},
		}, &buildResp); err != nil {
			return fmt.Errorf("unable to build component: %w", err)
		}

		// TODO(nnnnat): this is probably not the best way to access the build ID from this interface
		var buildID string = buildResp.JsonOutput.(map[string]interface{})["id"].(string)
		var releaseResp activities.CLICommandResponse
		if err := w.defaultExecGetActivity(ctx, w.acts.CLICommand, &activities.CLICommandRequest{
			OrgID:   outputs.OrgID,
			AppID:   outputs.AppID,
			Install: false,
			Json:    false,
			Args: []string{
				"releases",
				"create",
				"-c",
				compID,
				"-b",
				buildID,
			},
		}, &releaseResp); err != nil {
			return fmt.Errorf("unable to build component: %w", err)
		}
	}

	return nil
}

func (w *wkflow) execCLICommands(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	methods := []func(workflow.Context, *activities.TerraformRunOutputs) error{
		w.getCurrentOrg,
		w.getApp,
		w.listApps,
		w.listComponents,
		w.getComponent,
		w.listInstalls,
		w.getInstall,
		w.buildAndReleaseComponents,
	}

	fmt.Println("hello?")

	for _, method := range methods {
		if err := method(ctx, outputs); err != nil {
			return fmt.Errorf("error on cli command: %w", err)
		}
	}

	return nil
}
