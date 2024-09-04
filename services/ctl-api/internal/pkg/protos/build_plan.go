package protos

import (
	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) ToBuildPlanRequest(compBuild *app.ComponentBuild, compCfg *componentsv1.Component) *planv1.CreatePlanRequest {
	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     compBuild.OrgID,
				AppId:     compBuild.ComponentConfigConnection.Component.AppID,
				BuildId:   compBuild.ID,
				Component: compCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_BUILD,
				Context:   c.BuildContext(),
			},
		},
	}
}
