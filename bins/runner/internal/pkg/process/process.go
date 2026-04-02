package process

import (
	"context"
	"fmt"

	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/log"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
)

type Params struct {
	fx.In

	APIClient  nuonrunner.Client
	Cfg        *internal.Config
	L          *zap.Logger `name:"system"`
	LC         fx.Lifecycle
	Settings   *settings.Settings
	Shutdowner fx.Shutdowner
	Process    string `name:"process"`
}

type Registrar struct {
	processID   string
	processType string
	apiClient   nuonrunner.Client
	l           *zap.Logger
	cfg         *internal.Config
	settings    *settings.Settings
	shutdowner  fx.Shutdowner

	logStreamID string
	logProvider *otellog.LoggerProvider
	processLog  *zap.Logger
}

func New(params Params) (*Registrar, error) {
	processType := "install"
	if params.Process == "mng" {
		processType = "mng"
	}

	r := &Registrar{
		processType: processType,
		apiClient:   params.APIClient,
		l:           params.L,
		cfg:         params.Cfg,
		settings:    params.Settings,
		shutdowner:  params.Shutdowner,
	}

	params.LC.Append(r.lifecycleHook())
	return r, nil
}

func (r *Registrar) ProcessID() string {
	return r.processID
}

func (r *Registrar) LogStreamID() string {
	return r.logStreamID
}

// Logger returns the process-level logger that writes to the process log stream.
// Returns nil if the log stream was not initialized.
func (r *Registrar) Logger() *zap.Logger {
	return r.processLog
}

func (r *Registrar) lifecycleHook() fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			return r.start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return r.stop(ctx)
		},
	}
}

func (r *Registrar) start(ctx context.Context) error {
	r.l.Info("registering runner process",
		zap.String("type", r.processType),
		zap.String("version", r.cfg.Version),
	)

	process, err := r.apiClient.CreateProcess(ctx, &models.ServiceCreateRunnerProcessRequest{
		Type:    r.processType,
		Version: r.cfg.Version,
	})
	if err != nil {
		return fmt.Errorf("unable to create runner process: %w", err)
	}

	r.processID = process.ID
	r.logStreamID = process.LogStreamID

	r.l.Info("runner process registered",
		zap.String("process_id", r.processID),
		zap.String("type", r.processType),
		zap.String("log_stream_id", r.logStreamID),
	)

	// Set up process-level log stream
	if r.logStreamID != "" {
		lp, err := slog.NewOTELProvider(r.cfg, r.settings, r.logStreamID)
		if err != nil {
			r.l.Warn("unable to create process log provider", zap.Error(err))
		} else {
			r.logProvider = lp
			pl, err := log.NewOTELJobLogger(r.cfg, lp)
			if err != nil {
				r.l.Warn("unable to create process logger", zap.Error(err))
			} else {
				r.processLog = pl.With(
					zap.String("runner_process.id", r.processID),
					zap.String("runner_process.type", r.processType),
				)
				r.processLog.Info("process log stream initialized")
			}
		}
	}

	return nil
}

func (r *Registrar) stop(ctx context.Context) error {
	if r.processID == "" {
		return nil
	}

	if r.processLog != nil {
		r.processLog.Info("process shutting down")
	}

	r.l.Info("updating runner process status to shut-down",
		zap.String("process_id", r.processID),
	)

	status := "shut-down"
	_, err := r.apiClient.UpdateProcess(ctx, r.processID, &models.ServiceUpdateRunnerProcessRequest{
		Status:            &status,
		StatusDescription: "process stopped",
	})
	if err != nil {
		r.l.Warn("unable to update runner process status on shutdown",
			zap.String("process_id", r.processID),
			zap.Error(err),
		)
	}

	// Flush process log stream
	if r.logProvider != nil {
		if err := r.logProvider.ForceFlush(ctx); err != nil {
			r.l.Warn("unable to flush process log provider", zap.Error(err))
		}
	}

	return nil
}
