package plan

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"google.golang.org/protobuf/proto"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func FetchPlan(ctx context.Context, apiClient nuonrunner.Client, job *models.AppRunnerJob) (*planv1.Plan, error) {
	apiPlan, err := apiClient.GetJobPlan(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", err)
	}

	plan, err := apiPlanToProto(apiPlan)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api plan: %w", err)
	}

	return plan, nil
}

func apiPlanToProto(apiPlan *models.Planv1Plan) (*planv1.Plan, error) {
	planByts, err := apiPlan.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("unable to convert plan from API to byts: %w", err)
	}

	plan := &planv1.Plan{}
	if err := proto.Unmarshal(planByts, plan); err != nil {
		return nil, fmt.Errorf("unable to unmarshal plan bytes into proto: %w", err)
	}

	return plan, nil
}
