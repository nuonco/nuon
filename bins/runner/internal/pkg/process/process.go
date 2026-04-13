package process

import (
	"context"
	"fmt"
	"time"

	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/retry"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
	"github.com/nuonco/nuon/bins/runner/internal/version"
)

type Params struct {
	fx.In

	APIClient  nuonrunner.Client
	Cfg        *internal.Config
	LC         fx.Lifecycle
	Settings   *settings.Settings
	Shutdowner fx.Shutdowner
	Process    string `name:"process"`
}

type Result struct {
	fx.Out

	Registrar          *Registrar
	ProcessLogProvider *otellog.LoggerProvider `name:"process-log-provider" optional:"true"`
}

type Registrar struct {
	processID   string
	processType string
	apiClient   nuonrunner.Client
	cfg         *internal.Config
	settings    *settings.Settings
	shutdowner  fx.Shutdowner

	logStreamID string
	logProvider *otellog.LoggerProvider
}

// New creates a process and registers it with the API during initialization
// (not in a lifecycle hook) so that the process ID, log stream ID, and OTEL
// provider are available to other FX dependencies during their New() calls.
func New(params Params) (Result, error) {
	r := &Registrar{
		processType: params.Process,
		apiClient:   params.APIClient,
		cfg:         params.Cfg,
		settings:    params.Settings,
		shutdowner:  params.Shutdowner,
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	var process *models.AppRunnerProcess
	createFn := func(ctx context.Context) error {
		var err error
		process, err = r.apiClient.CreateProcess(ctx, &models.ServiceCreateRunnerProcessRequest{
			Type:    &r.processType,
			Version: version.Version,
		})
		return err
	}

	if err := retry.Retry(ctx, createFn, retry.WithMaxAttempts(3), retry.WithSleep(time.Second)); err != nil {
		return Result{}, fmt.Errorf("unable to create runner process: %w", err)
	}

	r.processID = process.ID
	r.logStreamID = process.LogStreamID

	if r.logStreamID != "" {
		lp, err := slog.NewOTELProvider(r.cfg, r.settings, r.logStreamID)
		if err == nil {
			r.logProvider = lp
		}
	}

	params.LC.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return r.stop(ctx)
		},
	})

	return Result{
		Registrar:          r,
		ProcessLogProvider: r.logProvider,
	}, nil
}

func (r *Registrar) ProcessID() string {
	return r.processID
}

func (r *Registrar) LogStreamID() string {
	return r.logStreamID
}

// LogProvider returns the OTEL log provider for the process log stream.
func (r *Registrar) LogProvider() *otellog.LoggerProvider {
	return r.logProvider
}

func (r *Registrar) stop(ctx context.Context) error {
	if r.processID == "" {
		return nil
	}

	status := "shut-down"
	_, err := r.apiClient.UpdateProcess(ctx, r.processID, &models.ServiceUpdateRunnerProcessRequest{
		Status:            &status,
		StatusDescription: "process stopped",
	})
	if err != nil {
		return fmt.Errorf("unable to update runner process status on shutdown: %w", err)
	}

	if r.logProvider != nil {
		if err := r.logProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("unable to flush process log provider: %w", err)
		}
	}

	return nil
}
