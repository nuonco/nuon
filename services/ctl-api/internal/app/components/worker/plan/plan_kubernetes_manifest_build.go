package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createKubernetesManifestBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.KubernetesManifestBuildPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	inputs, err := buildsource.ComponentInputs(&bld.ComponentConfigConnection)
	if err != nil {
		return nil, err
	}
	plan := inputs.KubernetesManifestBuildPlan
	plan.Labels = map[string]string{
		"component_id":       bld.ComponentID,
		"component_build_id": bld.ID,
	}

	if plan.SourceType == "kustomize" {
		l.Info("generating kustomize build plan")
	} else {
		l.Info("generating inline manifest build plan")
	}

	return plan, nil
}
