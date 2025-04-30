package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) getKubeClusterInfo(ctx workflow.Context, stack *app.InstallStack, state *state.State) (*kube.ClusterInfo, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	obj := &kube.ClusterInfo{
		ID:       "{{.nuon.sandbox.outputs.cluster.name}}",
		Endpoint: "{{.nuon.sandbox.outputs.cluster.endpoint}}",
		CAData:   "{{.nuon.sandbox.outputs.cluster.certificate_authority_data}}",
		AWSAuth: &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				RoleARN:     stack.InstallStackOutputs.AWSStackOutputs.MaintenanceIAMRoleARN,
				SessionName: "maintenance",
			},
		},
	}

	stateData, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state data")
	}

	if err := render.RenderStruct(obj, stateData); err != nil {
		l.Error("error rendering cluster info",
			zap.Any("cluster-info", obj),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	return obj, nil
}
