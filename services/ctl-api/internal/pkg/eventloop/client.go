package eventloop

import (
	"context"

	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

// TODO: make event loop status consts
type Params struct {
	fx.In

	V      *validator.Validate
	L      *zap.Logger
	Client temporalclient.Client
	MW     metrics.Writer
	Cfg    *internal.Config
	DB     *gorm.DB `name:"psql"`
}

type Client interface {
	Send(ctx context.Context, id string, signal Signal)
	GetWorkflowStatus(ctx context.Context, namespace string, workflowId string) (enumsv1.WorkflowExecutionStatus, error)
	GetWorkflowCount(ctx context.Context, namespace string, workflowId string) (int64, error)
	// GetNamespaceClient(ctx context.Context, namespace string) tclient.Client
}

var _ Client = (*evClient)(nil)

type evClient struct {
	l      *zap.Logger
	client temporalclient.Client
	mw     metrics.Writer
	cfg    *internal.Config
	db     *gorm.DB
	v      *validator.Validate
}

func New(params Params) Client {
	return &evClient{
		client: params.Client,
		l:      params.L,
		cfg:    params.Cfg,
		mw:     params.MW,
		db:     params.DB,
		v:      params.V,
	}
}
