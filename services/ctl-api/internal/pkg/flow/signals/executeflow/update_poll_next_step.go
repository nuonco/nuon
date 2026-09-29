package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type PollNextStepResponse struct {
	StepID  string     `json:"step_id"`
	StepIdx int        `json:"step_idx"`
	Status  app.Status `json:"status"`
}

func (s *Signal) pollNextStepHandler(ctx workflow.Context) (*PollNextStepResponse, error) {
	defer s.beginReadOnlyUpdate()()

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return nil, err
	}

	for _, step := range steps {
		switch step.Status.Status {
		case app.StatusSuccess, app.StatusDiscarded, app.StatusUserSkipped, app.StatusAutoSkipped:
			continue
		default:
			return &PollNextStepResponse{
				StepID:  step.ID,
				StepIdx: step.Idx,
				Status:  step.Status.Status,
			}, nil
		}
	}

	return &PollNextStepResponse{}, nil
}
