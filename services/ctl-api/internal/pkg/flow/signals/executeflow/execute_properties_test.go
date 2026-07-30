package executeflow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type pendingGroupPositionResult struct {
	Position        int
	Runnable        bool
	InitialDispatch bool
}

func firstPendingGroupPositionPropertyWorkflow(ctx workflow.Context) (pendingGroupPositionResult, error) {
	sig := Signal{WorkflowID: "property-flow"}
	decision := sig.residentScheduleDecision(ctx)
	_, initialDispatch := residentInitialGroupPosition(decision)
	return pendingGroupPositionResult{
		Position:        decision.Position,
		Runnable:        decision.State == residentScheduleRunnable,
		InitialDispatch: initialDispatch,
	}, nil
}

func TestResidentAwaitRetryGroupIsNeverRunnable(t *testing.T) {
	t.Run("parked group blocks pending siblings", hegel.Case(func(ht *hegel.T) {
		groupCount := hegel.Draw(ht, hegel.Integers(1, 6))
		blockedPosition := hegel.Draw(ht, hegel.Integers(0, groupCount-1))
		groupIndexBase := hegel.Draw(ht, hegel.Integers(0, 20))
		groupIndexStride := hegel.Draw(ht, hegel.Integers(1, 5))
		pendingSiblingCount := hegel.Draw(ht, hegel.Integers(1, 5))

		groups := make([]app.WorkflowStepGroup, groupCount)
		steps := make([]app.WorkflowStep, 0, groupCount+pendingSiblingCount)
		for position := range groupCount {
			groupIndex := groupIndexBase + position*groupIndexStride
			groupStatus := app.StatusSuccess
			if position == blockedPosition {
				groupStatus = app.StatusFailedPendingRetry
			} else if position > blockedPosition {
				groupStatus = app.StatusPending
			}
			groups[position] = app.WorkflowStepGroup{
				ID:       fmt.Sprintf("group-%d", position),
				GroupIdx: groupIndex,
				Status:   app.CompositeStatus{Status: groupStatus},
			}

			if position < blockedPosition {
				steps = append(steps, app.WorkflowStep{
					ID:       fmt.Sprintf("completed-%d", position),
					GroupIdx: groupIndex,
					Status:   app.CompositeStatus{Status: app.StatusSuccess},
				})
				continue
			}
			if position > blockedPosition {
				steps = append(steps, app.WorkflowStep{
					ID:       fmt.Sprintf("later-%d", position),
					GroupIdx: groupIndex,
					Status:   app.CompositeStatus{Status: app.StatusPending},
				})
				continue
			}

			steps = append(steps, app.WorkflowStep{
				ID:              "parked-plan",
				GroupIdx:        groupIndex,
				Status:          app.CompositeStatus{Status: app.StatusError},
				ResultDirective: string(directive.StepAwaitRetry),
			})
			for sibling := range pendingSiblingCount {
				steps = append(steps, app.WorkflowStep{
					ID:       fmt.Sprintf("pending-sibling-%d", sibling),
					GroupIdx: groupIndex,
					Status:   app.CompositeStatus{Status: app.StatusPending},
				})
			}
		}

		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment()
		activities := &workflowactivities.Activities{}
		env.OnActivity(
			activities.PkgWorkflowsFlowGetFlowStepGroups,
			mock.Anything,
			"property-flow",
		).Return(groups, nil).Once()
		env.OnActivity(
			activities.PkgWorkflowsFlowGetFlowSteps,
			mock.Anything,
			workflowactivities.GetFlowStepsRequest{FlowID: "property-flow"},
		).Return(steps, nil).Once()

		env.ExecuteWorkflow(firstPendingGroupPositionPropertyWorkflow)
		if err := env.GetWorkflowError(); err != nil {
			ht.Fatalf("workflow failed: %v", err)
		}

		var result pendingGroupPositionResult
		if err := env.GetWorkflowResult(&result); err != nil {
			ht.Fatalf("unable to read workflow result: %v", err)
		}
		if result.Runnable {
			ht.Fatalf(
				"parked group became runnable: blocked_position=%d returned_position=%d group_count=%d pending_siblings=%d",
				blockedPosition,
				result.Position,
				groupCount,
				pendingSiblingCount,
			)
		}
		if result.InitialDispatch {
			ht.Fatalf("blocked resident group was dispatched during cold start")
		}
	}, hegel.WithTestCases(100)))
}
