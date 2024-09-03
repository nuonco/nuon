package eventloop

import (
	"context"

	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

// TODO: make event loop status consts

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
}

func New(v *validator.Validate,
	l *zap.Logger,
	client temporalclient.Client,
	mw metrics.Writer,
	cfg *internal.Config,
	db *gorm.DB,
) Client {
	return &evClient{
		client: client,
		l:      l,
		cfg:    cfg,
		mw:     mw,
		db:     db,
	}
}
