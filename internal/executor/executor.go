package executor

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint/v2/pkg/client"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/fetch"
	"github.com/powertoolsdev/workers-executors/internal/task/poll"
	"github.com/powertoolsdev/workers-executors/internal/task/queue"
	"github.com/powertoolsdev/workers-executors/internal/task/upsert"
	"github.com/powertoolsdev/workers-executors/internal/task/validate"
	"github.com/powertoolsdev/workers-executors/internal/writer"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type clientProvider interface {
	GetClient(context.Context) (pb.WaypointClient, error)
	Close() error
}

type executor struct {
	// Provider clientProvider  `validate:"required"`
	Plan   *planv1.PlanRef `validate:"required"`
	Logger *zap.Logger     `validate:"required"`

	// internal state
	v        *validator.Validate
	client   pb.WaypointClient
	provider clientProvider
	f        *os.File
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
//	return func(e *executor) error {
//		e.Provider = p
//		return nil
//	}
// }

func WithPlan(p *planv1.PlanRef) executorOption {
	return func(e *executor) error {
		e.Plan = p
		return nil
	}
}

func WithLogger(l *zap.Logger) executorOption {
	return func(e *executor) error {
		e.Logger = l
		return nil
	}
}

func (e *executor) Execute(ctx context.Context) (*executev1.ExecutePlanResponse, error) {
	defer e.cleanup()
	resp := &executev1.ExecutePlanResponse{}
	e.Logger = e.Logger.With(zap.Any("plan_ref", e.Plan))

	e.Logger.Debug("fetching waypoint plan")
	bp, err := e.fetchWaypointPlan(ctx)
	if err != nil {
		return resp, err
	}

	e.Logger.Debug("getting waypoint client")
	if err = e.getClient(ctx, bp); err != nil {
		return resp, err
	}

	e.Logger.Debug("upserting waypoint project")
	if err = e.upsert(ctx, bp); err != nil {
		return resp, err
	}

	e.Logger.Debug("queuing waypoint job")
	jobID, err := e.queueJob(ctx, bp)
	if err != nil {
		return resp, err
	}

	e.Logger = e.Logger.With(zap.String("jobID", jobID))

	e.Logger.Debug("validating waypoint job")
	if err = e.validateJob(ctx, jobID); err != nil {
		return resp, err
	}

	e.Logger.Debug("creating temp file")
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("deployments-job-event-%s", jobID))
	if err != nil {
		return resp, err
	}
	e.f = tmpFile

	e.Logger.Debug("polling waypoint job")
	if err = e.pollJob(ctx, jobID); err != nil {
		return resp, err
	}

	e.Logger.Debug("uploading logs")
	if err = e.upload(ctx, bp); err != nil {
		return resp, err
	}

	return resp, nil
}

func (e *executor) cleanup() {
	if e.provider != nil {
		e.provider.Close()
	}

	if e.f != nil {
		name := e.f.Name()
		e.f.Close()
		if err := os.Remove(name); err != nil {
			e.Logger.Warn("failed to remove temp file", zap.String("file", name))
		}
	}
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

func (e *executor) getClient(ctx context.Context, bp *planv1.WaypointPlan) error {
	prov, err := client.NewOrgProvider(e.v, client.WithOrgConfig(client.Config{
		Address: bp.WaypointServer.Address,
		Token: client.Token{
			Namespace: bp.WaypointServer.TokenSecretNamespace,
			Name:      bp.WaypointServer.TokenSecretName,
		},
	}))
	if err != nil {
		return err
	}
	e.provider = prov

	c, err := prov.GetClient(ctx)
	if err != nil {
		return err
	}
	e.client = c

	return nil
}

func (e *executor) upsert(ctx context.Context, bp *planv1.WaypointPlan) error {
	upserter, err := upsert.New(
		e.v,
		upsert.WithClient(e.client),
		upsert.WithName(bp.WaypointRef.App),
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
	fw, err := writer.NewFile(e.v, writer.WithFile(e.f))
	if err != nil {
		return err
	}

	lw, err := writer.NewLog(e.v, writer.WithLog(e.Logger))
	if err != nil {
		return err
	}

	mw := writer.NewMultiWriter(fw, lw)

	poller, err := poll.New(
		e.v,
		poll.WithClient(e.client),
		poll.WithJobID(jobID),
		poll.WithWriter(mw),
	)
	if err != nil {
		return err
	}

	return poller.Poll(ctx)
}

func (e *executor) upload(ctx context.Context, bp *planv1.WaypointPlan) error {
	uploader := uploader.NewS3Uploader(
		bp.Outputs.Bucket,
		bp.Outputs.BucketPrefix,
		uploader.WithAssumeRoleARN(bp.Outputs.BucketAssumeRoleArn),
	)

	// NOTE(jdt): this is necessary as we need the real filename (w/o path) of the tmp file for uploading
	stat, err := e.f.Stat()
	if err != nil {
		return err
	}

	return uploader.UploadFile(ctx, os.TempDir(), stat.Name(), bp.Outputs.EventsKey)
}
