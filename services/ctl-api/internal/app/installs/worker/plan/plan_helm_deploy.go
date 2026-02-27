package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
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

//go:embed fake_helm_plan.json
var FakeHelmPlanJSON string

//go:embed fake_helm_plan_display.json
var FakeHelmPlanDisplayJSON string

func (p *Planner) createHelmDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
) (*plantypes.HelmDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	// parse out various config fields
	cfg := compBuild.ComponentConfigConnection.HelmComponentConfig
	if err := render.RenderStruct(cfg, stateData); err != nil {
		l.Error("error rendering helm config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	namespace := cfg.Namespace.ValueOrDefault("{{.nuon.install.id}}")
	renderedNamespace, err := render.RenderV2(namespace, stateData)
	if err != nil {
		l.Error("error rendering namespace",
			zap.String("namespace", namespace),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	driver := cfg.StorageDriver.ValueOrDefault("configmap")
	renderedDriver, err := render.RenderV2(driver, stateData)
	if err != nil {
		l.Error("error rendering driver",
			zap.String("driver", driver),
			zap.Error(err))

		return nil, errors.Wrap(err, "unable to render driver")
	}

	var helmChartID string
	if driver == "nuon" {
		hc, err := activities.AwaitGetHelmChartByOwnerID(ctx, installDeploy.InstallComponent.ID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get helm chart")
		}
		helmChartID = hc.ID
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
	}

	valuesFiles := []string(cfg.ValuesFiles)
	values := make([]plantypes.HelmValue, 0)
	for k, v := range generics.ToStringMap(cfg.Values) {
		v, err = render.RenderV2(v, stateData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to render")
		}

		values = append(values, plantypes.HelmValue{
			Name:  k,
			Value: v,
		})
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

		// Set auth on cluster info if present
		if clusterInfo != nil {
			clusterInfo.WithAWSAuth(awsAuth)
			clusterInfo.WithAzureAuth(azureAuth)
		}
	}

	return &plantypes.HelmDeployPlan{
		Name:            cfg.ChartName,
		Namespace:       renderedNamespace,
		CreateNamespace: true,
		StorageDriver:   renderedDriver,
		HelmChartID:     helmChartID,
		ValuesFiles:     valuesFiles,
		Values:          values,
		TakeOwnership:   cfg.TakeOwnership,

		ClusterInfo: clusterInfo,
		AWSAuth:     awsAuth,
		AzureAuth:   azureAuth,
	}, nil
}

func (p *Planner) createHelmDeploySandboxMode(ctx workflow.Context, req *plantypes.HelmDeployPlan) *plantypes.HelmSandboxMode {
	return &plantypes.HelmSandboxMode{
		PlanContents:        FakeHelmPlanJSON,
		PlanDisplayContents: FakeHelmPlanDisplayJSON,
	}
}
