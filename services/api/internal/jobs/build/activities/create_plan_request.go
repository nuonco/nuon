package activities

import (
	"context"
	"fmt"

	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/api/internal/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

type activities struct {
	db *gorm.DB
}

func New(db *gorm.DB) *activities {
	return &activities{db: db}
}

func (a *activities) CreatePlanRequest(ctx context.Context, req *apibuildv1.StartBuildRequest) (*planv1.CreatePlanRequest, error) {
	// TODO fix up logger here
	componentService := services.NewComponentService(a.db, zap.L())
	component, err := componentService.GetComponent(ctx, req.ComponentId)

	if err != nil {
		return nil, fmt.Errorf("error during StartBuild.CreatePlan loading component: %w", err)
	}

	componentProto, err := converters.ComponentModelToProto(component)
	if err != nil {
		return nil, fmt.Errorf("error during StartBuild.CreatePlan component ToProto: %w", err)
	}
	// create plan for container build
	planReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				Type:  planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_BUILD,
				OrgId: component.App.OrgID,
				AppId: component.AppID,
				// TODO adjust these structs accordingly. No deploymentid relevant to this anymore.
				DeploymentId: "dplfakeforcevalidation1234", // I think this can be empty
				Component: &componentv1.Component{
					Id:        component.ID,
					BuildCfg:  componentProto.ComponentConfig.BuildCfg,
					DeployCfg: componentProto.ComponentConfig.DeployCfg,
				},
			},
		},
	}

	return planReq, nil
}
