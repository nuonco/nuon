package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	tclient "go.temporal.io/sdk/client"
)

const (
	defaultNamespace                string        = "api"
	defaultWorkflowRunTimeout       time.Duration = time.Hour * 12
	defaultWorkflowExecutionTimeout time.Duration = time.Hour * 24
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_manager.go -source=manager.go -package=jobs
type Manager interface {
	CreateOrg(context.Context, string) error
	CreateApp(context.Context, string) error
	CreateInstall(context.Context, string) error
	CreateDeployment(context.Context, string) error
	StartDeploy(context.Context, string) (string, error)
}

var _ Manager = (*manager)(nil)

type manager struct {
	v *validator.Validate

	Namespace string                       `validate:"required"`
	Opts      tclient.StartWorkflowOptions `validate:"required"`
	Client    temporal.Client              `validate:"required"`
}

// New returns a default manager with the default orgcontext getter
func New(v *validator.Validate, opts ...managerOption) (*manager, error) {
	r := &manager{
		v:         v,
		Namespace: defaultNamespace,
		Opts: tclient.StartWorkflowOptions{
			WorkflowExecutionTimeout: defaultWorkflowExecutionTimeout,
			WorkflowRunTimeout:       defaultWorkflowRunTimeout,
			TaskQueue:                wfc.APITaskQueue,
			Memo: map[string]interface{}{
				"started-by": "default",
			},
		},
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate manager: %w", err)
	}

	return r, nil
}

type managerOption func(*manager) error

func WithClient(client temporal.Client) managerOption {
	return func(r *manager) error {
		r.Client = client
		return nil
	}
}
