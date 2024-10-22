package planv1

func FakeWaypointPlan() *Plan {
	return &Plan{
		Actual: &Plan_WaypointPlan{},
	}
}

func FakePlanResponse() *CreatePlanResponse {
	return &CreatePlanResponse{
		Plan: &Plan{
			Actual: &Plan_WaypointPlan{},
		},
		Ref: &PlanRef{},
	}
}
