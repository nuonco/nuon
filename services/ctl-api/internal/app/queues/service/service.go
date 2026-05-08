package service

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	V              *validator.Validate
	Logger         *zap.Logger
	TemporalClient temporalclient.Client
	QueueClient    *queueclient.Client
}

type service struct {
	db             *gorm.DB
	cfg            *internal.Config
	v              *validator.Validate
	l              *zap.Logger
	temporalClient temporalclient.Client
	queueClient    *queueclient.Client
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db:             params.DB,
		cfg:            params.Cfg,
		v:              params.V,
		l:              params.Logger,
		temporalClient: params.TemporalClient,
		queueClient:    params.QueueClient,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// New list and get endpoints
	queues := api.Group("/v1/queues")
	{
		queues.GET("", s.ListQueues)
		queues.GET("/:queue_id", s.GetQueue)
	}

	// Existing queue detail endpoints
	queueDetail := api.Group("/v1/queues/:queue_id")
	{
		queueDetail.GET("/status", s.GetQueueStatus)
		queueDetail.GET("/signals", s.GetQueueSignals)
		queueDetail.GET("/signals/:signal_id", s.GetQueueSignal)
		queueDetail.GET("/signals/:signal_id/await", s.AwaitQueueSignal)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	queues := api.Group("/v1/queues")
	{
		queues.POST("/:queue_id/admin-restart", s.RestartQueue)
		queues.POST("/:queue_id/signals/:signal_id/admin-direct-execute", s.DirectExecuteSignal)
	}
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

// getQueue retrieves a queue by ID from the database
func (s *service) getQueue(ctx *gin.Context, queueID string) (*app.Queue, error) {
	var queue app.Queue
	res := s.db.WithContext(ctx).
		Where("id = ?", queueID).
		First(&queue)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get queue: %w", res.Error)
	}

	return &queue, nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
