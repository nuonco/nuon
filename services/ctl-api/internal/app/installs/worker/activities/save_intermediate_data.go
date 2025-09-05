package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SaveIntermediateDataRequest struct {
	InstallID   string `validate:"required"`
	RunnerJobID string `validate:"required"`
	PlanJSON    string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SaveIntermediateData(ctx context.Context, req *SaveIntermediateDataRequest) error {
	plan, err := apiPlanToProto([]byte(req.PlanJSON))
	if err != nil {
		return errors.Wrap(err, "unable to convert to regular plan")
	}

	id := plan.GetWaypointPlan().Variables.IntermediateData.AsMap()
	byts, err := json.Marshal(id)
	if err != nil {
		return errors.Wrap(err, "unable to convert to json")
	}

	obj := app.InstallIntermediateData{
		InstallID:            req.InstallID,
		IntermediateDataJSON: string(byts),
		RunnerJobID:          req.RunnerJobID,
	}
	res := a.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to create intermediate data")
	}

	return nil
}

func apiPlanToProto(byts []byte) (*planv1.Plan, error) {
	plan := &planv1.Plan{}
	if err := protojson.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("unable to unmarshal plan bytes into proto: %w", err)
	}

	return plan, nil
}
