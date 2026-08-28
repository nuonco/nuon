package plan

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	_ "embed"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

//go:embed fake_terraform_state.json
var FakeTerraformStateJSON string

//go:embed fake_terraform_plan_contents.json
var FakeTerraformPlanContents string

//go:embed fake_terraform_plan_display_contents.json
var FakeTerraformPlanDisplayContents string

func (p *Planner) createTerraformDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
	roleSelection *operationroles.RoleSelection,
) (*plantypes.TerraformDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	installComp, err := activities.AwaitGetInstallComponentByID(
		ctx,
		installDeploy.InstallComponentID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(
		ctx,
		installDeploy.ComponentBuildID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	return p.RenderTerraformDeployPlan(l, &RenderTerraformDeployPlanInput{
		Stack:         stack,
		State:         state,
		StateData:     stateData,
		InstallDeploy: installDeploy,
		CompBuild:     compBuild,
		WorkspaceID:   installComp.TerraformWorkspace.ID,
		RoleSelection: roleSelection,
		ResolveClusterInfo: func(cloudAuth *CloudAuth) (*kube.ClusterInfo, error) {
			return p.resolveKubernetesContext(ctx, &compBuild.ComponentConfigConnection, appCfg, stack, state, cloudAuth)
		},
	})
}

// RenderTerraformDeployPlanInput carries the already-loaded data a terraform
// deploy plan is rendered from. Connected install workflows populate it from
// Temporal activities; install-free callers (e.g. the customer-managed bundle compiler)
// can populate it from an app config and a virtual install topology.
type RenderTerraformDeployPlanInput struct {
	Stack         *app.InstallStack
	State         *state.State
	StateData     map[string]any
	InstallDeploy *app.InstallDeploy
	CompBuild     *app.ComponentBuild
	WorkspaceID   string
	RoleSelection *operationroles.RoleSelection

	// ResolveClusterInfo resolves the kubernetes cluster the component should
	// target, if any. Injected because the connected path resolves it via a
	// workflow SideEffect at this exact point in the plan render; install-free
	// callers can pass a wrapper around ResolveKubernetesContextFromData.
	ResolveClusterInfo func(*CloudAuth) (*kube.ClusterInfo, error)
}

// RenderTerraformDeployPlan is the pure core of createTerraformDeployPlan: it
// renders a terraform deploy plan from already-loaded inputs with no Temporal
// dependency. createTerraformDeployPlan must remain a thin wrapper around it
// so connected installs and install-free plan compilation cannot diverge.
func (p *Planner) RenderTerraformDeployPlan(
	l *zap.Logger,
	in *RenderTerraformDeployPlanInput,
) (*plantypes.TerraformDeployPlan, error) {
	// render cross-platform values
	cfg := in.CompBuild.ComponentConfigConnection.TerraformModuleComponentConfig
	if err := render.RenderStruct(cfg, in.StateData); err != nil {
		l.Error("error rendering terraform config",
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}
	vars := generics.ToStringMapAny(cfg.Variables)
	if err := render.RenderMap(&vars, in.StateData); err != nil {
		l.Error("error rendering vars",
			zap.Any("vars", vars),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)

	if err := render.RenderMap(&envVars, in.StateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	// Install-level Terraform vars override, carried via a reserved synthetic
	// input. Appended as the final var-file so it wins over the vendor's vars map
	// and var_files (last -var-file wins). Empty is a no-op.
	varsFiles := []string(cfg.VariablesFiles)
	tfVarsOverride, err := p.installComponentOverride(
		in.State, in.StateData,
		config.TFVarsOverrideInputName(in.InstallDeploy.ComponentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render terraform vars override")
	}
	if tfVarsOverride != "" {
		varsFiles = append(varsFiles, tfVarsOverride)
	}

	cloudAuth, err := p.AuthForDeploy(
		l,
		in.RoleSelection,
		in.Stack,
		fmt.Sprintf("component-deploy-%s", in.InstallDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := in.ResolveClusterInfo(cloudAuth)
	if err != nil {
		l.Warn("unable to resolve kubernetes context, this usually means this was not a kubernetes application")
	}

	// construct plan from rendered values
	return &plantypes.TerraformDeployPlan{
		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: varsFiles,
		State:     in.State,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: in.WorkspaceID,
		},
		AzureAuth:   cloudAuth.Azure,
		AWSAuth:     cloudAuth.AWS,
		GCPAuth:     cloudAuth.GCP,
		ClusterInfo: clusterInfo,
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: false,
		},
	}, nil
}

func (p *Planner) createTerraformDeploySandboxMode(
	req *plantypes.TerraformDeployPlan,
) (*plantypes.TerraformSandboxMode, error) {
	pdcJSONByts := new(bytes.Buffer)
	if err := json.Compact(pdcJSONByts, []byte(FakeTerraformPlanDisplayContents)); err != nil {
		return nil, errors.Wrap(err, "unable to get json")
	}

	stJSONByts := new(bytes.Buffer)
	if err := json.Compact(stJSONByts, []byte(FakeTerraformStateJSON)); err != nil {
		return nil, errors.Wrap(err, "unable to get json")
	}

	return &plantypes.TerraformSandboxMode{
		WorkspaceID: req.TerraformBackend.WorkspaceID,

		StateJSON:           stJSONByts.Bytes(),
		PlanContents:        FakeTerraformPlanContents,
		PlanDisplayContents: pdcJSONByts.String(),
	}, nil
}
