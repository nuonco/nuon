package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createKubernetesManifestBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.KubernetesManifestBuildPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	cfg := bld.ComponentConfigConnection.KubernetesManifestComponentConfig
	if cfg == nil {
		return nil, errors.New("kubernetes manifest component config is nil")
	}

	plan := &plantypes.KubernetesManifestBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
	}

	switch {
	case cfg.Kustomize != nil && cfg.Kustomize.Path != "":
		l.Info("generating kustomize build plan")
		plan.SourceType = "kustomize"
		plan.KustomizePath = cfg.Kustomize.Path
		plan.KustomizeConfig = &plantypes.KustomizeBuildConfig{
			Patches:        cfg.Kustomize.Patches,
			EnableHelm:     cfg.Kustomize.EnableHelm,
			LoadRestrictor: cfg.Kustomize.LoadRestrictor,
		}

	default:
		l.Info("generating inline manifest build plan")
		plan.SourceType = "inline"
		plan.InlineManifest = cfg.Manifest
	}

	return plan, nil
}
