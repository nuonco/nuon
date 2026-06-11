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
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
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

	// render cross-platform values
	cfg := compBuild.ComponentConfigConnection.TerraformModuleComponentConfig
	if err := render.RenderStruct(cfg, stateData); err != nil {
		l.Error("error rendering terraform config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}
	vars := generics.ToStringMapAny(cfg.Variables)
	if err := render.RenderMap(&vars, stateData); err != nil {
		l.Error("error rendering vars",
			zap.Any("vars", vars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	envVars := generics.ToStringMap(cfg.EnvVars)

	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	// Install-level Terraform vars override, carried via a reserved synthetic
	// input. Appended as the final var-file so it wins over the vendor's vars map
	// and var_files (last -var-file wins). Empty is a no-op.
	varsFiles := []string(cfg.VariablesFiles)
	tfVarsOverride, err := p.installComponentOverride(
		state, stateData,
		config.TFVarsOverrideInputName(installDeploy.ComponentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render terraform vars override")
	}
	if tfVarsOverride != "" {
		varsFiles = append(varsFiles, tfVarsOverride)
	}

	cloudAuth, err := p.getAuthForDeploy(
		ctx,
		installDeploy,
		compBuild,
		appCfg,
		stack,
		state,
		fmt.Sprintf("component-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
	if err != nil {
		l.Warn("unable to get cluster information, this usually means this was not a kubernetes application")
	}

	// construct plan from rendered values
	return &plantypes.TerraformDeployPlan{
		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: varsFiles,
		State:     state,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: installComp.TerraformWorkspace.ID,
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
	ctx workflow.Context,
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
