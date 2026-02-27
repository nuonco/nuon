package plan

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	_ "embed"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
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
) (*plantypes.TerraformDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install id")
	}

	installComp, err := activities.AwaitGetInstallComponentByID(ctx, installDeploy.InstallComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

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

	envVars := generics.ToStringMap(cfg.EnvVars)

	if err := render.RenderMap(&envVars, stateData); err != nil {
		l.Error("error rendering env-vars",
			zap.Any("env-vars", envVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

	// Perform role selection for component deploys
	var awsAuth *awscredentials.Config
	var azureAuth *azurecredentials.Config

	if !org.SandboxMode {
		// Determine operation type based on deploy type
		var operation app.OperationType
		switch installDeploy.Type {
		case app.InstallDeployTypeApply:
			operation = app.OperationDeploy
		case app.InstallDeployTypeTeardown:
			operation = app.OperationTeardown
		default:
			operation = app.OperationDeploy
		}

		// Get default role from app permissions config
		// Components use MaintenanceRole for deploy and teardown operations
		defaultRole := appCfg.PermissionsConfig.MaintenanceRole.Name

		// Build selection context
		// TODO: Add component-specific operation roles when available
		selectionCtx := &operationroles.SelectionContext{
			Operation:     operation,
			PrincipalType: principal.TypeComponent,
			RuntimeRole:   installDeploy.Role,
			EntityRoles:   nil, // Components don't currently have entity-specific operation roles
			MatrixRules:   appCfg.OperationRoleConfig.Rules,
			DefaultRole:   defaultRole,
			AppConfig:     appCfg,
			StackOutputs:  &stack.InstallStackOutputs,
			InstallState:  state,
		}

		// Select role using operation roles engine
		roleSelection, err := operationroles.SelectRole(selectionCtx, l)
		if err != nil {
			return nil, errors.Wrap(err, "unable to select role for deploy")
		}

		l.Info("selected role for component deploy",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
			zap.String("source", string(roleSelection.Source)),
		)

		// Create auth configuration with selected role
		awsAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("component-deploy-%s", installDeploy.ID),
				RoleARN:     roleSelection.RoleARN,
			},
		}
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
