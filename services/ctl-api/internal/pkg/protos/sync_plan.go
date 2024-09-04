package protos

import (
	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) ToSyncPlanRequest(install *app.Install, installDeploy *app.InstallDeploy, deployCfg *componentsv1.Component) *planv1.CreatePlanRequest {
	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     install.App.OrgID,
				AppId:     install.App.ID,
				BuildId:   installDeploy.ComponentBuildID,
				InstallId: install.ID,
				DeployId:  installDeploy.ID,
				Component: deployCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_SYNC_IMAGE,
				Context:   c.InstallContext(install),
			},
		},
	}
}
