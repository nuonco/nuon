package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

//go:embed fake_helm_plan.json
var FakeHelmPlanJSON string

//go:embed fake_helm_plan_display.json
var FakeHelmPlanDisplayJSON string

func (p *Planner) createHelmDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.HelmDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
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

	driver := cfg.StorageDriver.ValueOrDefault("secrets")
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
	}, nil
}

func (p *Planner) createHelmDeploySandboxMode(ctx workflow.Context, req *plantypes.HelmDeployPlan) *plantypes.HelmSandboxMode {
	return &plantypes.HelmSandboxMode{
		PlanContents:        FakeHelmPlanJSON,
		PlanDisplayContents: FakeHelmPlanDisplayJSON,
	}
}
