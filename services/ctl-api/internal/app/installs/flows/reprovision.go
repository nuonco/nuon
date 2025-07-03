package flows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func Reprovision(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	step, err := installSignalStep(ctx, installID, "reprovision runner service account", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionRunner,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateInstallStackVersion,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitInstallStackVersionRun,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = installSignalStep(ctx, installID, "update install stack outputs", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationUpdateInstallStackOutputs,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = installSignalStep(ctx, installID, "await runner health", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreReprovision)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = installSignalStep(ctx, installID, "reprovision sandbox plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionSandboxPlan,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)
	step, err = installSignalStep(ctx, installID, "reprovision sandbox apply plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionSandboxApplyPlan,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreSecretsSync)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = installSignalStep(ctx, installID, "sync secrets", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationSyncSecrets,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostSecretsSync)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = installSignalStep(ctx, installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationProvisionDNS,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	deploySteps, err := deployAllComponents(ctx, installID, flw)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostReprovision)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
