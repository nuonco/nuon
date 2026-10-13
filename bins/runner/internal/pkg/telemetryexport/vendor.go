package telemetryexport

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/telemetryexport"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type VendorParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Settings  *settings.Settings
	Logger    *zap.Logger `name:"system"`
	APIClient nuonrunner.Client
	Audit     *audit.Writer
}

// VendorSupervisor owns the shared supervisor's run context for the FX lifecycle.
type VendorSupervisor struct {
	supervisor *telemetryexport.VendorSupervisor
	cancel     context.CancelFunc
	done       chan struct{}
}

func NewVendor(params VendorParams) *VendorSupervisor {
	return newVendor(params, telemetryexport.DefaultConfig(collectorStorageDir))
}

func newVendor(params VendorParams, config telemetryexport.Config) *VendorSupervisor {
	installID := params.Settings.Metadata["install.id"]
	if params.Settings.Cfg.IsNuonctl {
		// Local runners never export vendor telemetry; an empty install ID makes
		// Run remove stale relay credentials without fetching API settings.
		installID = ""
	}
	s := &VendorSupervisor{
		supervisor: telemetryexport.NewVendor(telemetryexport.Options{
			InstallID: installID,
			Binary:    collectorBinary,
			Config:    config,
			Client:    params.APIClient,
			Logger:    params.Logger,
			InitialSettings: &telemetryexport.Settings{
				Enabled:            params.Settings.VendorTelemetryEnabled,
				RelayEndpoint:      params.Settings.TelemetryRelayEndpoint,
				ResourceAttributes: params.Settings.VendorTelemetryResourceAttributes,
			},
			FetchSettings: func(ctx context.Context) (telemetryexport.Settings, error) {
				return fetchVendorSettings(ctx, params.APIClient)
			},
			OnEnabled: vendorTelemetryAudit(params.Audit, params.Logger),
		}),
	}
	params.Lifecycle.Append(fx.Hook{OnStart: s.start, OnStop: s.stop})
	return s
}

func (s *VendorSupervisor) start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})
	go func() {
		defer close(s.done)
		s.supervisor.Run(ctx)
	}()
	return nil
}

func (s *VendorSupervisor) stop(ctx context.Context) error {
	if s.cancel == nil {
		return nil
	}
	s.cancel()
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("vendor telemetry export supervisor did not stop: %w", ctx.Err())
	}
}

func fetchVendorSettings(ctx context.Context, client nuonrunner.Client) (telemetryexport.Settings, error) {
	response, err := client.GetSettings(ctx)
	if err != nil {
		return telemetryexport.Settings{}, err
	}
	if response == nil {
		return telemetryexport.Settings{}, fmt.Errorf("runner settings response is empty")
	}
	return telemetryexport.Settings{
		Enabled:            response.VendorTelemetryEnabled,
		RelayEndpoint:      response.TelemetryRelayEndpoint,
		ResourceAttributes: response.VendorTelemetryResourceAttributes,
	}, nil
}

type auditAsyncWriter interface {
	WriteAsync(audit.Event) error
}

func vendorTelemetryAudit(writer auditAsyncWriter, logger *zap.Logger) func(bool) {
	return func(enabled bool) {
		event := audit.Event{
			Name:       "install_telemetry_updated",
			Message:    "runner applied telemetry setting",
			Outcome:    audit.OutcomeSucceeded,
			Attributes: map[string]string{"telemetry.enabled": strconv.FormatBool(enabled)},
		}
		if err := writer.WriteAsync(event); err != nil && !errors.Is(err, audit.ErrUnavailable) {
			logger.Warn("customer telemetry audit event enqueue failed", zap.Error(err))
		}
	}
}
