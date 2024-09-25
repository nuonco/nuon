package plan

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"google.golang.org/protobuf/encoding/protojson"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func FetchPlan(ctx context.Context, apiClient nuonrunner.Client, job *models.AppRunnerJob) (*planv1.Plan, error) {
	planJSON, err := apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", err)
	}

	planByts := []byte(planJSON)

	plan, err := apiPlanToProto(planByts)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api plan: %w", err)
	}

	return plan, nil
}

func apiPlanToProto(byts []byte) (*planv1.Plan, error) {
	plan := &planv1.Plan{}
	if err := protojson.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("unable to unmarshal plan bytes into proto: %w", err)
	}

	return plan, nil
}
