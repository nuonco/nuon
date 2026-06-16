package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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
	roleSelection *operationroles.RoleSelection,
) (*plantypes.HelmDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
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
		hc, err := activities.AwaitGetHelmChartByOwnerID(
			ctx,
			installDeploy.InstallComponent.ID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get helm chart")
		}
		helmChartID = hc.ID
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

	// Install-level Helm values override, carried via a reserved synthetic input.
	// Rendered like app values so it can reference {{.nuon.*}}. Empty is a no-op.
	valuesOverride, err := p.installComponentOverride(
		state, stateData,
		config.HelmValuesOverrideInputName(installDeploy.ComponentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render helm values override")
	}

	cloudAuth, err := p.getAuthForDeploy(
		ctx,
		roleSelection,
		stack,
		fmt.Sprintf("component-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
	}

	return &plantypes.HelmDeployPlan{
		Name:            cfg.ChartName,
		Namespace:       renderedNamespace,
		CreateNamespace: true,
		StorageDriver:   renderedDriver,
		HelmChartID:     helmChartID,
		ValuesFiles:     valuesFiles,
		Values:          values,
		ValuesOverride:  valuesOverride,
		TakeOwnership:   cfg.TakeOwnership,

		ClusterInfo: clusterInfo,
		AWSAuth:     cloudAuth.AWS,
		AzureAuth:   cloudAuth.Azure,
		GCPAuth:     cloudAuth.GCP,
	}, nil
}

func (p *Planner) createHelmDeploySandboxMode(
	ctx workflow.Context,
	req *plantypes.HelmDeployPlan,
) *plantypes.HelmSandboxMode {
	return &plantypes.HelmSandboxMode{
		PlanContents:        FakeHelmPlanJSON,
		PlanDisplayContents: FakeHelmPlanDisplayJSON,
	}
}
