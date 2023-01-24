package executor

import (
	"context"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-waypoint/v2/pkg/client"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/fetch"
	"github.com/powertoolsdev/workers-executors/internal/task/poll"
	"github.com/powertoolsdev/workers-executors/internal/task/queue"
	"github.com/powertoolsdev/workers-executors/internal/task/upsert"
	"github.com/powertoolsdev/workers-executors/internal/task/validate"
	"github.com/powertoolsdev/workers-executors/internal/writer"
	"google.golang.org/protobuf/proto"
)

// type clientProvider interface {
// 	GetClient(context.Context) (pb.WaypointClient, error)
// 	Close() error
// }

type executor struct {
	// Provider clientProvider  `validate:"required"`
	Plan *planv1.PlanRef `validate:"required:dive"`

	// internal state
	v      *validator.Validate
	client pb.WaypointClient
}

type executorOption func(*executor) error

func New(v *validator.Validate, opts ...executorOption) (*executor, error) {
	e := &executor{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating executor: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	if err := e.v.Struct(e); err != nil {
		return nil, err
	}

	return e, nil
}

// func WithProvider(p clientProvider) executorOption {
// 	return func(e *executor) error {
// 		e.Provider = p
// 		return nil
// 	}
// }

func WithPlan(p *planv1.PlanRef) executorOption {
	return func(e *executor) error {
		e.Plan = p
		return nil
	}
}

func (e *executor) Execute(ctx context.Context) (interface{}, error) {
	bp, err := e.fetchWaypointPlan(ctx)
	if err != nil {
		return nil, err
	}

	cleanup, err := e.getClient(ctx, bp)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cleanup() }()

	if err = e.upsert(ctx, bp); err != nil {
		return nil, err
	}

	jobID, err := e.queueJob(ctx, bp)
	if err != nil {
		return nil, err
	}

	if err = e.validateJob(ctx, jobID); err != nil {
		return nil, err
	}

	if err = e.pollJob(ctx, jobID); err != nil {
		return nil, err
	}

	return nil, nil
}

func (e *executor) fetchWaypointPlan(ctx context.Context) (*planv1.WaypointPlan, error) {
	fetcher, err := fetch.NewS3(
		e.v,
		fetch.WithBucketKey(e.Plan.BucketKey),
		fetch.WithBucketName(e.Plan.Bucket),
		fetch.WithRoleARN(e.Plan.BucketAssumeRoleArn),
		fetch.WithRoleSessionName("fetch-waypoint-plan"),
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

	var bp planv1.WaypointPlan
	if err = proto.Unmarshal(bs, &bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

func (e *executor) getClient(ctx context.Context, bp *planv1.WaypointPlan) (func() error, error) {
	prov, err := client.NewOrgProvider(e.v, client.WithOrgConfig(client.Config{
		Address: bp.WaypointServer.Address,
		Token: client.Token{
			Namespace: bp.WaypointServer.TokenSecretNamespace,
			Name:      bp.WaypointServer.TokenSecretName,
		},
	}))
	if err != nil {
		return nil, err
	}

	c, err := prov.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	e.client = c

	return prov.Close, nil
}

func (e *executor) upsert(ctx context.Context, bp *planv1.WaypointPlan) error {
	upserter, err := upsert.New(
		e.v,
		upsert.WithClient(e.client),
		upsert.WithName(bp.Component.Name),
		upsert.WithProject(bp.WaypointRef.Project),
	)

	if err != nil {
		return err
	}

	return upserter.UpsertWaypointApplication(ctx)
}

func (e *executor) queueJob(ctx context.Context, bp *planv1.WaypointPlan) (string, error) {
	queuer, err := queue.New(
		e.v,
		queue.WithClient(e.client),
		queue.WithID(bp.WaypointRef.SingletonId),
		queue.WithWorkspace(bp.WaypointRef.Workspace),
		queue.WithApplication(bp.WaypointRef.App),
		queue.WithProject(bp.WaypointRef.Project),
		queue.WithLabels(bp.WaypointRef.Labels),
		queue.WithGitURL("https://github.com/jonmorehouse/empty"), // TODO(jdt): should this come from the plan?
		queue.WithWaypointHCL([]byte(bp.WaypointRef.HclConfig)),   // TODO(jdt): we probably shouldn't be going back and forth...
		queue.WithTargetRunnerID(bp.WaypointRef.RunnerId),
		queue.WithOnDemandRunnerName(bp.WaypointRef.OnDemandRunnerConfig),
	)
	if err != nil {
		return "", err
	}

	return queuer.QueueDeployment(ctx)
}

func (e *executor) validateJob(ctx context.Context, jobID string) error {
	validator, err := validate.New(e.v, validate.WithClient(e.client), validate.WithID(jobID))
	if err != nil {
		return err
	}

	return validator.Validate(ctx)
}

func (e *executor) pollJob(ctx context.Context, jobID string) error {
	nopWriter, err := writer.NewNop(e.v)
	if err != nil {
		return err
	}

	poller, err := poll.New(
		e.v,
		poll.WithClient(e.client),
		poll.WithJobID(jobID),
		poll.WithWriter(nopWriter),
	)
	if err != nil {
		return err
	}

	return poller.Poll(ctx)
}
