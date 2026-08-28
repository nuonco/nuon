package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

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

	return p.RenderHelmDeployPlan(l, &RenderHelmDeployPlanInput{
		Stack:         stack,
		State:         state,
		StateData:     stateData,
		InstallDeploy: installDeploy,
		CompBuild:     compBuild,
		RoleSelection: roleSelection,
		GetHelmChartID: func(ownerID string) (string, error) {
			hc, err := activities.AwaitGetHelmChartByOwnerID(ctx, ownerID)
			if err != nil {
				return "", err
			}
			return hc.ID, nil
		},
		ResolveClusterInfo: func(cloudAuth *CloudAuth) (*kube.ClusterInfo, error) {
			return p.resolveKubernetesContext(ctx, &compBuild.ComponentConfigConnection, appCfg, stack, state, cloudAuth)
		},
	})
}

// RenderHelmDeployPlanInput carries the already-loaded data a helm deploy plan
// is rendered from.
type RenderHelmDeployPlanInput struct {
	Stack         *app.InstallStack
	State         *state.State
	StateData     map[string]any
	InstallDeploy *app.InstallDeploy
	CompBuild     *app.ComponentBuild
	RoleSelection *operationroles.RoleSelection

	// GetHelmChartID loads the chart at this exact point in the plan render.
	GetHelmChartID func(string) (string, error)
	// ResolveClusterInfo resolves the kubernetes cluster at this exact point in
	// the plan render.
	ResolveClusterInfo func(*CloudAuth) (*kube.ClusterInfo, error)
}

// RenderHelmDeployPlan is the pure core of createHelmDeployPlan.
func (p *Planner) RenderHelmDeployPlan(
	l *zap.Logger,
	in *RenderHelmDeployPlanInput,
) (*plantypes.HelmDeployPlan, error) {
	// parse out various config fields
	cfg := in.CompBuild.ComponentConfigConnection.HelmComponentConfig
	if err := render.RenderStruct(cfg, in.StateData); err != nil {
		l.Error("error rendering helm config",
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	namespace := cfg.Namespace.ValueOrDefault("{{.nuon.install.id}}")
	renderedNamespace, err := render.RenderV2(namespace, in.StateData)
	if err != nil {
		l.Error("error rendering namespace",
			zap.String("namespace", namespace),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	driver := cfg.StorageDriver.ValueOrDefault("configmap")
	renderedDriver, err := render.RenderV2(driver, in.StateData)
	if err != nil {
		l.Error("error rendering driver",
			zap.String("driver", driver),
			zap.Error(err))

		return nil, errors.Wrap(err, "unable to render driver")
	}

	var helmChartID string
	if driver == "nuon" {
		helmChartID, err = in.GetHelmChartID(in.InstallDeploy.InstallComponent.ID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get helm chart")
		}
	}

	valuesFiles := []string(cfg.ValuesFiles)
	values := make([]plantypes.HelmValue, 0)
	for k, v := range generics.ToStringMap(cfg.Values) {
		v, err = render.RenderV2(v, in.StateData)
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
		in.State, in.StateData,
		config.HelmValuesOverrideInputName(in.InstallDeploy.ComponentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render helm values override")
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
		return nil, errors.Wrap(err, "unable to resolve kubernetes context")
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
		SkipCRDs:        cfg.SkipCRDs,

		ClusterInfo: clusterInfo,
		AWSAuth:     cloudAuth.AWS,
		AzureAuth:   cloudAuth.Azure,
		GCPAuth:     cloudAuth.GCP,
	}, nil
}

func (p *Planner) createHelmDeploySandboxMode(
	req *plantypes.HelmDeployPlan,
) *plantypes.HelmSandboxMode {
	return &plantypes.HelmSandboxMode{
		PlanContents:        FakeHelmPlanJSON,
		PlanDisplayContents: FakeHelmPlanDisplayJSON,
	}
}
