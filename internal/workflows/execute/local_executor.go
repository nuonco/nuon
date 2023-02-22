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
	"google.golang.org/protobuf/types/known/structpb"
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
	output, err := e.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("executor did not succeed: %w", err)
	}

	resp, err := a.parseOutput(l, plan, output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse planner outputs: %w", err)
	}

	return resp, nil
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
		l.Debug("got waypoint plan", "plan", plan)
		zapLog := zap.L()
		return waypoint.New(a.v, waypoint.WithPlan(actual.WaypointPlan), waypoint.WithLogger(zapLog))

	case *planv1.Plan_TerraformPlan:
		l.Debug("got terraform plan", "plan", plan)
		return terraform.New(a.v, terraform.WithPlan(actual.TerraformPlan))

	default:
		typ := fmt.Sprintf("%T", actual)
		l.Debug("got unknown plan", "plan", plan, "type", typ)
		return nil, fmt.Errorf("unknown plan type: %s", typ)
	}
}

func (a *Activities) parseOutput(l log.Logger, plan *planv1.Plan, output interface{}) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}

	switch actual := plan.Actual.(type) {
	case *planv1.Plan_WaypointPlan:
		l.Debug("waypoint plan not parsing outputs")
		return resp, nil

	case *planv1.Plan_TerraformPlan:
		l.Debug("parsing outputs for terraform plan")
		m, ok := output.(map[string]interface{})
		if !ok {
			l.Warn("terraform output not correct type")
			return resp, fmt.Errorf("terraform output is incorrect type: %T", output)
		}
		s, err := structpb.NewStruct(m)
		if err != nil {
			return resp, err
		}
		resp.Outputs = &executev1.ExecutePlanResponse_TerraformOutputs{TerraformOutputs: s}

		return resp, nil

	default:
		typ := fmt.Sprintf("%T", actual)
		l.Debug("got unknown plan", "plan", plan, "type", typ)
		return nil, fmt.Errorf("unknown plan type: %s", typ)
	}
}
