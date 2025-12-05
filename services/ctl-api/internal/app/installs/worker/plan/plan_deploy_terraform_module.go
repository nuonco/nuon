package plan

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	_ "embed"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

//go:embed fake_terraform_state.json
var FakeTerraformStateJSON string

//go:embed fake_terraform_plan_contents.json
var FakeTerraformPlanContents string

//go:embed fake_terraform_plan_display_contents.json
var FakeTerraformPlanDisplayContents string

func (p *Planner) createTerraformDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.TerraformDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install id")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	installComp, err := activities.AwaitGetInstallComponentByID(ctx, installDeploy.InstallComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
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

	var clusterInfo *kube.ClusterInfo
	if !org.SandboxMode {
		clusterInfo, err = p.getKubeClusterInfo(ctx, stack, state)
		if err != nil {
			l.Warn("unable to get cluster information, this usually means this was not a kubernetes application")
		}
	}

	// render platform-specific values
	var awsAuth *awscredentials.Config
	var azureAuth *azurecredentials.Config
	envVars := generics.ToStringMap(cfg.EnvVars)
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		roleARN := stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN
		awsAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-deploy-%s", req.InstallDeployID),
				RoleARN:     roleARN,
			},
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		azureAuth = &azurecredentials.Config{
			UseDefault: true,
		}
		envVars["ARM_SUBSCRIPTION_ID"] = "{{.nuon.install_stack.outputs.subscription_id}}"
	}
	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	// construct plan from rendered values
	return &plantypes.TerraformDeployPlan{
		Vars:      vars,
		EnvVars:   envVars,
		VarsFiles: cfg.VariablesFiles,
		State:     state,

		TerraformBackend: &plantypes.TerraformBackend{
			WorkspaceID: installComp.TerraformWorkspace.ID,
		},
		AzureAuth:   azureAuth,
		AWSAuth:     awsAuth,
		ClusterInfo: clusterInfo,
		Hooks: &plantypes.TerraformDeployHooks{
			Enabled: false,
		},
	}, nil
}

func (p *Planner) createTerraformDeploySandboxMode(ctx workflow.Context, req *plantypes.TerraformDeployPlan) (*plantypes.TerraformSandboxMode, error) {
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
