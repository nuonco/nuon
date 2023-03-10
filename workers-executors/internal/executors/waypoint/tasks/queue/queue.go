package queue

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"google.golang.org/grpc"
)

const (
	defaultJobTimeout string = "1h"
)

type jobQueuer interface {
	QueueJob(ctx context.Context, in *gen.QueueJobRequest, opts ...grpc.CallOption) (*gen.QueueJobResponse, error)
	GetLatestPushedArtifact(ctx context.Context, in *gen.GetLatestPushedArtifactRequest, opts ...grpc.CallOption) (*gen.PushedArtifact, error)
}

var _ jobQueuer = (gen.WaypointClient)(nil)

// TODO(jdt): robust-ify validation?
type queuer struct {
	Client             jobQueuer              `validate:"required"`
	ID                 string                 `validate:"required"`
	Workspace          string                 `validate:"required"`
	Project            string                 `validate:"required"`
	Application        string                 `validate:"required"`
	Labels             map[string]string      `validate:"required"`
	WaypointHCL        []byte                 `validate:"required"`
	TargetRunnerID     string                 `validate:"required"`
	OnDemandRunnerName string                 `validate:"required"`
	JobTimeout         string                 `validate:"required"`
	JobType            planv1.WaypointJobType `validate:"required"`
	GitURL             string                 `validate:"required"`
	Path               string
	CommitRef          string

	// internal state
	v *validator.Validate
}

type queuerOption func(*queuer) error

func New(v *validator.Validate, opts ...queuerOption) (*queuer, error) {
	q := &queuer{v: v, JobTimeout: defaultJobTimeout}

	if v == nil {
		return nil, fmt.Errorf("error instantiating executor: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(q); err != nil {
			return nil, err
		}
	}

	if err := q.v.Struct(q); err != nil {
		return nil, err
	}

	return q, nil
}

func WithClient(c jobQueuer) queuerOption {
	return func(q *queuer) error {
		q.Client = c
		return nil
	}
}

func WithID(id string) queuerOption {
	return func(q *queuer) error {
		q.ID = id
		return nil
	}
}

func WithWorkspace(ws string) queuerOption {
	return func(q *queuer) error {
		q.Workspace = ws
		return nil
	}
}

func WithApplication(id string) queuerOption {
	return func(q *queuer) error {
		q.Application = id
		return nil
	}
}

func WithProject(id string) queuerOption {
	return func(q *queuer) error {
		q.Project = id
		return nil
	}
}

func WithLabels(labels map[string]string) queuerOption {
	return func(q *queuer) error {
		q.Labels = labels
		return nil
	}
}

func WithGitURL(url string) queuerOption {
	return func(q *queuer) error {
		q.GitURL = url
		return nil
	}
}

func WithPath(path string) queuerOption {
	return func(q *queuer) error {
		q.Path = path
		return nil
	}
}

func WithCommitRef(ref string) queuerOption {
	return func(q *queuer) error {
		q.CommitRef = ref
		return nil
	}
}

func WithWaypointHCL(hcl []byte) queuerOption {
	return func(q *queuer) error {
		q.WaypointHCL = hcl
		return nil
	}
}

func WithTargetRunnerID(id string) queuerOption {
	return func(q *queuer) error {
		q.TargetRunnerID = id
		return nil
	}
}

func WithOnDemandRunnerName(name string) queuerOption {
	return func(q *queuer) error {
		q.OnDemandRunnerName = name
		return nil
	}
}

func WithJobType(typ planv1.WaypointJobType) queuerOption {
	return func(q *queuer) error {
		q.JobType = typ
		return nil
	}
}

func WithJobTimeout(timeout string) queuerOption {
	return func(q *queuer) error {
		q.JobTimeout = timeout
		return nil
	}
}

func (q *queuer) getArtifact(ctx context.Context) (*gen.PushedArtifact, error) {
	push, err := q.Client.GetLatestPushedArtifact(ctx, &gen.GetLatestPushedArtifactRequest{
		Application: &gen.Ref_Application{
			Application: q.Application,
			Project:     q.Project,
		},
		Workspace: &gen.Ref_Workspace{
			Workspace: q.Workspace,
		},
	})
	if err != nil {
		return nil, err
	}

	return push, nil
}

// QueueDeployment queues the job returning the jobID or error
func (q *queuer) QueueDeployment(ctx context.Context) (string, error) {
	req := &gen.QueueJobRequest{
		Job: &gen.Job{
			SingletonId: q.ID,
			Workspace: &gen.Ref_Workspace{
				Workspace: q.Workspace,
			},
			Application: &gen.Ref_Application{
				Project:     q.Project,
				Application: q.Application,
			},
			Labels: q.Labels,
			DataSource: &gen.Job_DataSource{
				Source: &gen.Job_DataSource_Git{
					Git: &gen.Job_Git{
						Url:  q.GitURL,
						Path: q.Path,
						Ref:  q.CommitRef,
					},
				},
			},
			WaypointHcl: &gen.Hcl{
				Contents: q.WaypointHCL,
				// TODO(jdt): accept this from the plan
				// Format: q.Format,
			},
			TargetRunner: &gen.Ref_Runner{
				Target: &gen.Ref_Runner_Id{
					Id: &gen.Ref_RunnerId{
						Id: q.TargetRunnerID,
					},
				},
			},
			OndemandRunner: &gen.Ref_OnDemandRunnerConfig{
				Name: q.OnDemandRunnerName,
			},

			Variables: []*gen.Variable{},
		},
		ExpiresIn: q.JobTimeout,
	}

	switch q.JobType {
	case planv1.WaypointJobType_WAYPOINT_JOB_TYPE_BUILD:
		req.Job.Operation = &gen.Job_Build{
			Build: &gen.Job_BuildOp{DisablePush: false},
		}
	case planv1.WaypointJobType_WAYPOINT_JOB_TYPE_DEPLOY_ARTIFACT:
		artifact, err := q.getArtifact(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to get artifact: %w", err)
		}
		req.Job.Operation = &gen.Job_Deploy{
			Deploy: &gen.Job_DeployOp{
				Artifact: artifact,
			},
		}
	case planv1.WaypointJobType_WAYPOINT_JOB_TYPE_DEPLOY:
		req.Job.Operation = &gen.Job_Deploy{
			Deploy: &gen.Job_DeployOp{},
		}
	default:
		return "", fmt.Errorf("invalid job type: %s", q.JobType)
	}

	resp, err := q.Client.QueueJob(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.JobId, nil
}
