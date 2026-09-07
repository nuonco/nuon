package telemetryexport

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"

	"github.com/nuonco/nuon/pkg/runner/settings"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	vendorCollectorHealthURL       = "http://" + vendorCollectorHealthAddress + "/"
	vendorSettingsRefreshInterval  = 15 * time.Second
	vendorSettingsRequestTimeout   = 15 * time.Second
	vendorCollectorRestartInterval = time.Second
)

type vendorSettings struct {
	enabled  bool
	endpoint string
}

type VendorParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Settings  *settings.Settings
	Logger    *zap.Logger `name:"system"`
	APIClient nuonrunner.Client
}

type VendorSupervisor struct {
	installID           string
	initialSettings     vendorSettings
	activeEndpoint      string
	desiredEndpoint     string
	rejectedEndpoint    string
	local               bool
	logger              *zap.Logger
	tokens              tokenLifecycle
	cancel              context.CancelFunc
	done                chan struct{}
	mu                  sync.Mutex
	child               *childProcess
	backoff             time.Duration
	nextStart           time.Time
	enabled             bool
	disabled            bool
	settingsUnavailable bool

	fetchSettingsFn func(context.Context) (vendorSettings, error)
	replaceChildFn  func(string) error
	stopChildFn     func()
}

func NewVendor(params VendorParams) *VendorSupervisor {
	s := &VendorSupervisor{
		installID: params.Settings.Metadata["install.id"],
		initialSettings: vendorSettings{
			enabled:  params.Settings.VendorTelemetryEnabled,
			endpoint: params.Settings.TelemetryRelayEndpoint,
		},
		local:   params.Settings.Cfg.IsNuonctl,
		logger:  params.Logger,
		tokens:  newTokenManager(params.APIClient, params.Logger),
		done:    make(chan struct{}),
		backoff: time.Second,
	}
	s.fetchSettingsFn = func(ctx context.Context) (vendorSettings, error) {
		response, err := params.APIClient.GetSettings(ctx)
		if err != nil {
			return vendorSettings{}, err
		}
		if response == nil {
			return vendorSettings{}, fmt.Errorf("runner settings response is empty")
		}
		return vendorSettings{enabled: response.VendorTelemetryEnabled, endpoint: response.TelemetryRelayEndpoint}, nil
	}
	s.replaceChildFn = s.replaceChild
	s.stopChildFn = s.stopChild
	params.Lifecycle.Append(fx.Hook{OnStart: s.start, OnStop: s.stop})
	return s
}

func (s *VendorSupervisor) start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx)
	return nil
}

func (s *VendorSupervisor) stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
		select {
		case <-s.done:
		case <-ctx.Done():
			s.stopChildFn()
		}
	}
	s.tokens.Disable()
	return nil
}

func (s *VendorSupervisor) run(ctx context.Context) {
	defer close(s.done)
	if s.local || s.installID == "" {
		s.tokens.Disable()
		return
	}

	s.reconcile(ctx, s.initialSettings)
	settingsTicker := time.NewTicker(vendorSettingsRefreshInterval)
	restartTicker := time.NewTicker(vendorCollectorRestartInterval)
	defer settingsTicker.Stop()
	defer restartTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.stopChildFn()
			s.tokens.Disable()
			return
		case <-settingsTicker.C:
			s.refreshSettings(ctx)
		case <-restartTicker.C:
			s.restartIfNeeded(ctx)
		}
	}
}

func (s *VendorSupervisor) refreshSettings(ctx context.Context) {
	requestCtx, cancel := context.WithTimeout(ctx, vendorSettingsRequestTimeout)
	defer cancel()
	settings, err := s.fetchSettingsFn(requestCtx)
	if err != nil {
		if !s.settingsUnavailable {
			s.logger.Warn("vendor telemetry settings unavailable; retaining active configuration", zap.Error(err))
		}
		s.settingsUnavailable = true
		return
	}
	if s.settingsUnavailable {
		s.logger.Info("vendor telemetry settings available")
	}
	s.settingsUnavailable = false
	s.reconcile(ctx, settings)
}

func (s *VendorSupervisor) reconcile(ctx context.Context, settings vendorSettings) {
	if !settings.enabled || settings.endpoint == "" {
		s.disable()
		return
	}
	if err := validateOTLPHTTPExporter(otlpHTTPExporter{Endpoint: settings.endpoint}); err != nil {
		if settings.endpoint != s.rejectedEndpoint {
			s.logger.Warn("vendor telemetry relay endpoint is invalid; retaining active configuration", zap.Error(err))
		}
		if s.enabled {
			s.desiredEndpoint = s.activeEndpoint
			s.nextStart = time.Time{}
			s.backoff = time.Second
		} else {
			s.disable()
		}
		s.rejectedEndpoint = settings.endpoint
		return
	}
	s.rejectedEndpoint = ""
	if settings.endpoint == s.desiredEndpoint {
		return
	}

	s.desiredEndpoint = settings.endpoint
	s.disabled = false
	s.backoff = time.Second
	s.nextStart = time.Time{}
	if s.enabled && s.activeEndpoint == s.desiredEndpoint {
		return
	}
	s.startCollector(ctx)
}

func (s *VendorSupervisor) disable() {
	if s.disabled {
		return
	}
	s.stopChildFn()
	s.tokens.Disable()
	s.activeEndpoint = ""
	s.desiredEndpoint = ""
	s.rejectedEndpoint = ""
	s.nextStart = time.Time{}
	s.backoff = time.Second
	s.enabled = false
	s.disabled = true
	s.logger.Info("vendor telemetry export collector disabled")
}

func (s *VendorSupervisor) startCollector(ctx context.Context) {
	if err := s.tokens.Enable(ctx); err != nil {
		s.logger.Warn("vendor telemetry access token unavailable", zap.Error(err))
		s.scheduleRestart()
		return
	}

	previousEndpoint := s.activeEndpoint
	previousEnabled := s.enabled
	if err := s.replaceChildFn(s.desiredEndpoint); err != nil {
		s.logger.Warn("vendor telemetry export collector failed to start", zap.Error(err))
		if previousEnabled && previousEndpoint != "" && previousEndpoint != s.desiredEndpoint {
			if rollbackErr := s.replaceChildFn(previousEndpoint); rollbackErr == nil {
				s.activeEndpoint = previousEndpoint
				s.enabled = true
				s.scheduleRestart()
				return
			} else {
				s.logger.Error("vendor telemetry export collector rollback failed", zap.Error(rollbackErr))
			}
		}
		s.tokens.Disable()
		s.activeEndpoint = ""
		s.enabled = false
		s.scheduleRestart()
		return
	}

	s.enabled = true
	s.activeEndpoint = s.desiredEndpoint
	s.nextStart = time.Time{}
	endpoint, _ := url.Parse(s.activeEndpoint)
	s.logger.Info("vendor telemetry export collector enabled",
		zap.String("vendor_telemetry_export.backend", endpoint.Host),
		zap.Bool("vendor_telemetry_export.logs", true),
		zap.Bool("vendor_telemetry_export.metrics", true),
		zap.Bool("vendor_telemetry_export.traces", true),
	)
}

func (s *VendorSupervisor) replaceChild(endpoint string) error {
	if _, err := os.Stat(collectorBinary); err != nil {
		return err
	}
	contents, err := vendorCollectorConfig(endpoint)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp("", "nuon-vendor-telemetry-export-")
	if err != nil {
		return err
	}
	path := filepath.Join(tempDir, "collector.yaml")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		os.RemoveAll(tempDir)
		return err
	}
	cmd := exec.Command(collectorBinary, "--config", path)
	cmd.Env = childEnvironment(nil)
	stdout := &zapio.Writer{Log: s.logger.Named("vendor-telemetry-export-collector").With(zap.String("stream", "stdout")), Level: zapcore.WarnLevel}
	stderr := &zapio.Writer{Log: s.logger.Named("vendor-telemetry-export-collector").With(zap.String("stream", "stderr")), Level: zapcore.WarnLevel}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	s.stopChild()
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		os.RemoveAll(tempDir)
		return err
	}
	child := &childProcess{cmd: cmd, tempDir: tempDir, done: make(chan struct{}), startedAt: time.Now(), outputWriters: []io.Closer{stdout, stderr}}
	s.mu.Lock()
	s.child = child
	s.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		for _, writer := range child.outputWriters {
			_ = writer.Close()
		}
		close(child.done)
	}()
	if err := waitForCollector(child, vendorCollectorHealthURL); err != nil {
		s.stopChild()
		return err
	}
	return nil
}

func (s *VendorSupervisor) stopChild() {
	s.mu.Lock()
	child := s.child
	s.child = nil
	s.mu.Unlock()
	if child == nil {
		return
	}
	if child.cmd.Process != nil {
		_ = child.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-child.done:
		case <-time.After(5 * time.Second):
			_ = child.cmd.Process.Kill()
			<-child.done
		}
	}
	_ = os.RemoveAll(child.tempDir)
}

func (s *VendorSupervisor) restartIfNeeded(ctx context.Context) {
	s.mu.Lock()
	child := s.child
	s.mu.Unlock()
	if child != nil {
		select {
		case <-child.done:
			s.stopChildFn()
			s.enabled = false
			s.activeEndpoint = ""
			s.logger.Warn("vendor telemetry export collector exited; scheduling restart")
			if time.Since(child.startedAt) >= 30*time.Second {
				s.backoff = time.Second
			}
			s.scheduleRestart()
		default:
		}
	}
	if s.nextStart.IsZero() || time.Now().Before(s.nextStart) {
		return
	}
	s.startCollector(ctx)
}

func (s *VendorSupervisor) scheduleRestart() {
	s.nextStart = time.Now().Add(s.backoff)
	s.backoff *= 2
	if s.backoff > 30*time.Second {
		s.backoff = 30 * time.Second
	}
}
