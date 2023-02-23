package waypoint

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint/v2/pkg/client"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/tasks/poll"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/tasks/queue"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/tasks/upsert"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/tasks/validate"
	"github.com/powertoolsdev/workers-executors/internal/writer"
	"go.uber.org/zap"
)

type clientProvider interface {
	GetClient(context.Context) (pb.WaypointClient, error)
	Close() error
}

type executor struct {
	// Provider clientProvider  `validate:"required"`
	Plan   *planv1.WaypointPlan `validate:"required"`
	Logger *zap.Logger          `validate:"required"`

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

func WithPlan(p *planv1.WaypointPlan) executorOption {
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

func (e *executor) Execute(ctx context.Context) (interface{}, error) {
	defer e.cleanup()
	var resp interface{}

	e.Logger.Debug("getting waypoint client")
	if err := e.getClient(ctx); err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	e.Logger.Debug("upserting waypoint application")
	if err := e.upsert(ctx); err != nil {
		return resp, fmt.Errorf("unable to upsert application: %w", err)
	}

	e.Logger.Debug("queuing waypoint job")
	jobID, err := e.queueJob(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to queue job:% w", err)
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
	if err = e.upload(ctx); err != nil {
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

func (e *executor) getClient(ctx context.Context) error {
	prov, err := client.NewOrgProvider(e.v, client.WithOrgConfig(client.Config{
		Address: e.Plan.WaypointServer.Address,
		Token: client.Token{
			Namespace: e.Plan.WaypointServer.TokenSecretNamespace,
			Name:      e.Plan.WaypointServer.TokenSecretName,
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

func (e *executor) upsert(ctx context.Context) error {
	upserter, err := upsert.New(
		e.v,
		upsert.WithClient(e.client),
		upsert.WithName(e.Plan.WaypointRef.App),
		upsert.WithProject(e.Plan.WaypointRef.Project),
	)
	if err != nil {
		return err
	}

	return upserter.UpsertWaypointApplication(ctx)
}

func (e *executor) queueJob(ctx context.Context) (string, error) {
	queuer, err := queue.New(
		e.v,
		queue.WithClient(e.client),
		queue.WithID(e.Plan.WaypointRef.SingletonId),
		queue.WithWorkspace(e.Plan.WaypointRef.Workspace),
		queue.WithApplication(e.Plan.WaypointRef.App),
		queue.WithProject(e.Plan.WaypointRef.Project),
		queue.WithLabels(e.Plan.WaypointRef.Labels),
		queue.WithGitURL(e.Plan.GitSource.Url),
		queue.WithPath(e.Plan.GitSource.Path),
		queue.WithCommitRef(e.Plan.GitSource.Ref),
		queue.WithWaypointHCL([]byte(e.Plan.WaypointRef.HclConfig)),
		queue.WithTargetRunnerID(e.Plan.WaypointRef.RunnerId),
		queue.WithOnDemandRunnerName(e.Plan.WaypointRef.OnDemandRunnerConfig),
		queue.WithJobType(e.Plan.WaypointRef.JobType),
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

func (e *executor) upload(ctx context.Context) error {
	uploader, err := uploader.NewS3Uploader(
		e.v,
		uploader.WithBucketName(e.Plan.Outputs.Bucket),
		uploader.WithAssumeRoleARN(e.Plan.Outputs.BucketAssumeRoleArn),
		uploader.WithAssumeSessionName("workers-executors"),
	)
	if err != nil {
		return fmt.Errorf("unable to get uploader: %w", err)
	}

	// NOTE(jdt): this is necessary as we need the real filename (w/o path) of the tmp file for uploading
	stat, err := e.f.Stat()
	if err != nil {
		return err
	}

	return uploader.UploadFile(ctx, os.TempDir(), stat.Name(), e.Plan.Outputs.EventsKey)
}
