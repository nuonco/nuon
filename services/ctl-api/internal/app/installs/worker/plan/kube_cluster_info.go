package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (p *Planner) getKubeClusterInfo(ctx workflow.Context, stack *app.InstallStack, state *state.State) (*kube.ClusterInfo, error) {
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
		return nil, errors.Wrap(err, "unable to render config")
	}

	return obj, nil
}
