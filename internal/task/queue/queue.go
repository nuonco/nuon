package queue

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

const (
	defaultJobTimeout string = "1h"
)

type jobQueuer interface {
	QueueJob(ctx context.Context, in *gen.QueueJobRequest, opts ...grpc.CallOption) (*gen.QueueJobResponse, error)
}

var _ jobQueuer = (gen.WaypointClient)(nil)

// TODO(jdt): robust-ify validation?
type queuer struct {
	Client             jobQueuer         `validate:"required"`
	ID                 string            `validate:"required"`
	Workspace          string            `validate:"required"`
	Project            string            `validate:"required"`
	Application        string            `validate:"required"`
	Labels             map[string]string `validate:"required"`
	GitURL             string            `validate:"required"`
	WaypointHCL        []byte            `validate:"required"`
	TargetRunnerID     string            `validate:"required"`
	OnDemandRunnerName string            `validate:"required"`
	JobTimeout         string            `validate:"required"`

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

func WithJobTimeout(timeout string) queuerOption {
	return func(q *queuer) error {
		q.JobTimeout = timeout
		return nil
	}
}

// QueueDeployment queues the job returning the jobID or error
func (q *queuer) QueueDeployment(ctx context.Context) (string, error) {
	req := &gen.QueueJobRequest{
		Job: &gen.Job{
			Operation: &gen.Job_Build{
				Build: &gen.Job_BuildOp{DisablePush: false},
			},
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
						Url: q.GitURL,
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

	resp, err := q.Client.QueueJob(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.JobId, nil
}
