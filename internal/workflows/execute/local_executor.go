package execute

import (
	"context"
	"fmt"
	"io"

	s3fetch "github.com/powertoolsdev/go-fetch/pkg/s3"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/executors/terraform"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type executor interface {
	Execute(context.Context) (interface{}, error)
}

func (a *Activities) ExecutePlanLocally(ctx context.Context, req *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	l := log.With(activity.GetLogger(ctx), "plan", req.Plan)

	l.Debug("fetching plan")
	plan, err := a.fetchPlan(ctx, req.Plan)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan: %w", err)
	}

	l.Debug("determining executor")
	e, err := a.getExecutor(l, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to determine or instantiate executor: %w", err)
	}

	l.Debug("executing plan")
	_, err = e.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("executor did not succeed: %w", err)
	}
	return &executev1.ExecutePlanResponse{}, nil
}

func (a *Activities) fetchPlan(ctx context.Context, planref *planv1.PlanRef) (*planv1.Plan, error) {
	fetcher, err := s3fetch.New(
		a.v,
		s3fetch.WithBucketKey(planref.BucketKey),
		s3fetch.WithBucketName(planref.Bucket),
		s3fetch.WithRoleARN(planref.BucketAssumeRoleArn),
		s3fetch.WithRoleSessionName("workers-executors-fetch-plan"),
	)
	if err != nil {
		return nil, err
	}

	iorc, err := fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	defer iorc.Close()

	bs, err := io.ReadAll(iorc)
	if err != nil {
		return nil, err
	}

	var bp planv1.Plan
	if err = proto.Unmarshal(bs, &bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

func (a *Activities) getExecutor(l log.Logger, plan *planv1.Plan) (executor, error) {
	switch actual := plan.Actual.(type) {
	case *planv1.Plan_WaypointPlan:
		l.Debug("got waypoint plan", zap.Any("plan", plan))
		zapLog := zap.L()
		return waypoint.New(a.v, waypoint.WithPlan(actual.WaypointPlan), waypoint.WithLogger(zapLog))

	case *planv1.Plan_TerraformPlan:
		l.Debug("got terraform plan", zap.Any("plan", plan))
		return terraform.New(a.v, terraform.WithPlan(actual.TerraformPlan))

	default:
		typ := fmt.Sprintf("%T", actual)
		l.Debug("got unknown plan", zap.Any("plan", plan), zap.String("type", typ))
		return nil, fmt.Errorf("unknown plan type: %s", typ)
	}
}
