package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/distribution/reference"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createActionWorkflowRunPlan(ctx workflow.Context, runID string) (*plantypes.ActionWorkflowRunPlan, *operationroles.RoleSelection, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, nil, err
	}

	l.Info("creating plan for executing action workflow")
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, runID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get run")
	}

	slimInstall, err := activities.AwaitGetSlimInstallByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, slimInstall.AppConfigID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get app config")
	}

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, run.InstallID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, nil, errors.Wrap(err, "unable to get install state")
	}

	stateMap, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to convert state to map")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, run.InstallID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get install stack")
	}

	builtInEnvVars, err := p.getBuiltinEnvVars(ctx, run)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get env vars")
	}

	overrideEnvVars, err := p.getOverrideEnvVars(ctx, run)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get override env vars")
	}

	attrs := ActionWorkflowRunAttrs(run)

	cloudAuth, roleSelection, err := p.getAuthForActionWorkflowRun(ctx, stack.InstallStackOutputs, run, appCfg, stack, state)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get auth for action workflow run")
	}
	var clusterInfo *kube.ClusterInfo
	if run.EnableKubeConfig.Valid && run.EnableKubeConfig.Bool {
		// Target the action's declared kubernetes_context, falling through to
		// the sandbox default when it's empty (adhoc runs, or actions that
		// don't declare a context).
		clusterInfo, err = p.resolveKubernetesContextByName(ctx, run.KubernetesContextName, appCfg, stack, state, cloudAuth)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to resolve kubernetes context")
		}
	}

	plan, selectedRole, err := p.RenderActionWorkflowRunPlan(l, &RenderActionWorkflowRunPlanInput{
		RunID:           runID,
		Run:             run,
		SandboxMode:     slimInstall.SandboxMode.Bool,
		AppConfig:       appCfg,
		StateMap:        stateMap,
		BuiltinEnvVars:  builtInEnvVars,
		OverrideEnvVars: overrideEnvVars,
		Attrs:           attrs,
		CloudAuth:       cloudAuth,
		RoleSelection:   roleSelection,
		ClusterInfo:     clusterInfo,
		GetStepGitSource: func(stepID string) (*plantypes.GitSource, error) {
			return activities.AwaitGetActionWorkflowStepGitSourceByStepID(ctx, stepID)
		},
	})
	if err != nil {
		return nil, nil, err
	}

	if !run.ActionWorkflowConfigID.Empty() && run.ActionWorkflowConfig.Image != "" {
		sourceImage, err := RenderText(run.ActionWorkflowConfig.Image, stateMap)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to render action image")
		}

		if err := p.setActionImagePlan(ctx, plan, sourceImage, runID, stack, stateMap, cloudAuth); err != nil {
			return nil, nil, err
		}
	}

	return plan, selectedRole, nil
}

// RenderActionWorkflowRunPlanInput carries the already-loaded data an action workflow run plan is rendered from.
type RenderActionWorkflowRunPlanInput struct {
	RunID            string
	Run              *app.InstallActionWorkflowRun
	SandboxMode      bool
	AppConfig        *app.AppConfig
	StateMap         map[string]any
	BuiltinEnvVars   map[string]string
	OverrideEnvVars  map[string]string
	Attrs            map[string]string
	CloudAuth        *CloudAuth
	RoleSelection    *operationroles.RoleSelection
	ClusterInfo      *kube.ClusterInfo
	GetStepGitSource func(string) (*plantypes.GitSource, error)
}

// ActionWorkflowRunAttrs derives the run-level plan attributes from a run.
// Connected workflows call it before auth resolution (its original position);
// install-free callers use it to populate RenderActionWorkflowRunPlanInput.
func ActionWorkflowRunAttrs(run *app.InstallActionWorkflowRun) map[string]string {
	attrs := make(map[string]string, 0)
	if !run.ActionWorkflowConfigID.Empty() {
		attrs["action.name"] = run.ActionWorkflowConfig.ActionWorkflow.Name
		attrs["action.id"] = run.ActionWorkflowConfig.ActionWorkflow.ID
	} else {
		name := generics.FirstNonEmptyString(run.Steps[0].AdHocConfig.Name, "Adhoc Action")
		attrs["action.name"] = name
		attrs["action.id"] = run.ID
	}
	return attrs
}

// RenderActionWorkflowRunPlan renders an action workflow run plan with no Temporal dependency.
func (p *Planner) RenderActionWorkflowRunPlan(l *zap.Logger, in *RenderActionWorkflowRunPlanInput) (*plantypes.ActionWorkflowRunPlan, *operationroles.RoleSelection, error) {
	run := in.Run

	plan := &plantypes.ActionWorkflowRunPlan{
		InstallID:       run.InstallID,
		ID:              in.RunID,
		Steps:           make([]*plantypes.ActionWorkflowRunStepPlan, 0),
		BuiltinEnvVars:  in.BuiltinEnvVars,
		OverrideEnvVars: in.OverrideEnvVars,
		Attrs:           in.Attrs,
		ClusterInfo:     in.ClusterInfo,
		AzureAuth:       in.CloudAuth.Azure,
		AWSAuth:         in.CloudAuth.AWS,
		GCPAuth:         in.CloudAuth.GCP,
	}

	if !run.ActionWorkflowConfigID.Empty() {
		if run.ActionWorkflowConfig.Timeout > 0 {
			plan.Timeout = run.ActionWorkflowConfig.Timeout
		}
		for idx, stepCfg := range run.Steps {
			l.Debug(fmt.Sprintf("creating plan for step %d", idx))
			l.Debug("creating git source for config")
			gitSource, err := in.GetStepGitSource(stepCfg.Step.ID)
			if err != nil {
				l.Error("unable to  configure git source for step", zap.Error(err))
				return nil, nil, errors.Wrap(errors.Wrap(err, "unable to get git source"), fmt.Sprintf("unable to create plan for step %d", idx))
			}
			stepPlan, err := p.RenderActionWorkflowStepPlan(l, &stepCfg, in.StateMap, gitSource)
			if err != nil {
				return nil, nil, errors.Wrap(err, fmt.Sprintf("unable to create plan for step %d", idx))
			}

			plan.Steps = append(plan.Steps, stepPlan)
		}
	} else {
		if run.Timeout > 0 {
			plan.Timeout = run.Timeout
		}
		stepPlan, err := p.RenderActionWorkflowAdhocStepPlan(l, &run.Steps[0], in.StateMap)
		if err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("unable to create adhoc step plan"))
		}
		plan.Steps = append(plan.Steps, stepPlan)
	}

	if in.SandboxMode {
		targetRefs := helpers.GetActionReferences(in.AppConfig, run.ActionWorkflowConfig.ActionWorkflow.Name)

		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: refs.GetFakeRefs(targetRefs),
		}
	}

	l.Info("successfully created plan")
	return plan, in.RoleSelection, nil
}

func (p *Planner) setActionImagePlan(
	ctx workflow.Context,
	plan *plantypes.ActionWorkflowRunPlan,
	sourceImage string,
	runID string,
	stack *app.InstallStack,
	stateMap map[string]interface{},
	cloudAuth *CloudAuth,
) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	plan.SourceImage = sourceImage

	named, err := reference.ParseDockerRef(sourceImage)
	if err != nil {
		return fmt.Errorf("invalid action image reference %q: %w", sourceImage, err)
	}

	loginServer := installRegistryLoginServer(stateMap, stack)
	if loginServer != "" && reference.Domain(named) == loginServer {
		digested, ok := named.(reference.Digested)
		if !ok {
			return fmt.Errorf(
				"action image %q resolves to the install registry but is not digest-pinned; reference the component's image.ref output rather than repository and tag",
				sourceImage,
			)
		}

		registryCfg := getInstallRegistryPullConfig(
			reference.TrimNamed(named).String(),
			loginServer,
			stack,
			cloudAuth,
		)
		if registryCfg == nil {
			return fmt.Errorf("unable to build install registry config for action image %q", sourceImage)
		}

		plan.ImageRegistry = registryCfg
		plan.ImageDigestRef = sourceImage

		l.Info("action image resolved to the install registry, skipping mirror",
			zap.String("action.image", sourceImage),
			zap.String("image.digest", digested.Digest().String()),
		)

		return nil
	}

	imageRegistry, err := p.getOrgRegistryRepositoryConfig(ctx, plan.InstallID, runID)
	if err != nil {
		return errors.Wrap(err, "unable to get registry for action image")
	}

	plan.ImageRegistry = imageRegistry
	plan.ImageTag = actionImageTag(sourceImage, runID)

	return nil
}

func actionImageTag(sourceImage, runID string) string {
	sum := sha256.Sum256([]byte(sourceImage))
	return fmt.Sprintf("action-%s-%s", hex.EncodeToString(sum[:])[:16], runID)
}

// TODO(ja): make this a method on the run struct?
func hstoreToMap(hstore pgtype.Hstore) map[string]string {
	result := make(map[string]string)
	for key, value := range hstore {
		result[key] = *value
	}
	return result
}

func (p *Planner) getRoleForAction(
	ctx workflow.Context,
	l *zap.Logger,
	appCfg *app.AppConfig,
	run *app.InstallActionWorkflowRun,
	stack *app.InstallStack,
	installState *state.State,
) (*operationroles.RoleSelection, app.OperationType, error) {
	flw := p.installWorkflowForRoleDefault(ctx, l, generics.FromPtrStr(run.InstallWorkflowID))
	return operationroles.GetRoleForAction(l, appCfg, run, stack, installState, flw)
}

func (p *Planner) getAuthForActionWorkflowRun(
	ctx workflow.Context,
	outputs app.InstallStackOutputs,
	run *app.InstallActionWorkflowRun,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	installState *state.State,
) (*CloudAuth, *operationroles.RoleSelection, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, nil, err
	}

	roleSelection, operation, err := p.getRoleForAction(ctx, l, appCfg, run, stack, installState)
	if err != nil {
		return nil, nil, err
	}

	l.Info("selected role for action workflow run plan",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
		zap.String("run_id", run.ID),
	)

	cloudAuth, err := getCloudAuth(roleSelection, &outputs, fmt.Sprintf("action-workflow-%s", run.ID))
	return cloudAuth, roleSelection, err
}
