package planv1

import (
	structpb "google.golang.org/protobuf/types/known/structpb"

	variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"
)

func FakeWaypointPlan() *Plan {
	return &Plan{
		Actual: &Plan_WaypointPlan{},
	}
}

func FakePlanResponse() *CreatePlanResponse {
	intermediateData := map[string]any{
		"nuon": map[string]any{
			"org": map[string]any{
				"id": "org-id",
			},
			"install": map[string]any{
				"id": "app-id",
			},
			"app": map[string]any{
				"id": "app-id",
			},
		},
	}

	id, err := structpb.NewStruct(intermediateData)
	if err != nil {
		panic("sandbox mode error when creating fake intermediate data " + err.Error())
	}

	return &CreatePlanResponse{
		Plan: &Plan{
			Actual: &Plan_WaypointPlan{
				WaypointPlan: &WaypointPlan{
					Variables: &variablesv1.Variables{
						IntermediateData: id,
					},
				},
			},
		},
		Ref: &PlanRef{},
	}
}
