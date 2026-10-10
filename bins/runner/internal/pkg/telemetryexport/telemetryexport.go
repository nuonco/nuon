package telemetryexport

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/telemetryexport"
)

const (
	collectorBinary    = "/bin/nuon-runner-otelcol"
	collectorHealthURL = "http://127.0.0.1:13133/"
	secretSyncInterval = 30 * time.Second
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Settings  *settings.Settings
	Logger    *zap.Logger `name:"system"`
	Sources   configSourceResolver
}

type Supervisor struct {
	installID        string
	platform         string
	local            bool
	logger           *zap.Logger
	sources          configSourceResolver
	cancel           context.CancelFunc
	done             chan struct{}
	mu               sync.Mutex
	child            *telemetryexport.Collector
	active           string
	rejectedConfig   string
	restartConfig    string
	backoff          time.Duration
	nextStart        time.Time
	reported         bool
	collectorEnabled bool

	replaceChildFn func(context.Context, config) error
	stopChildFn    func()
}

func New(params Params) *Supervisor {
	s := &Supervisor{installID: params.Settings.Metadata["install.id"], platform: params.Settings.Platform, local: params.Settings.Cfg.IsNuonctl, logger: params.Logger, sources: params.Sources, done: make(chan struct{}), backoff: time.Second}
	s.replaceChildFn = s.replaceChild
	s.stopChildFn = s.stopChild
	params.Lifecycle.Append(fx.Hook{OnStart: s.start, OnStop: s.stop})
	return s
}

func (s *Supervisor) start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx)
	return nil
}

func (s *Supervisor) stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
		select {
		case <-s.done:
		case <-ctx.Done():
			return fmt.Errorf("audit telemetry export supervisor did not stop: %w", ctx.Err())
		}
	}
	return nil
}

func (s *Supervisor) run(ctx context.Context) {
	defer close(s.done)
	if s.local || s.installID == "" {
		return
	}
	source := s.sources.Resolve(s.platform, s.installID)
	if source == nil {
		return
	}
	defer s.stopChildFn()
	updates := source.Watch(ctx, secretSyncInterval)
	crash := time.NewTicker(time.Second)
	defer crash.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			s.reconcile(ctx, update)
		case <-crash.C:
			s.restartCrashed(ctx)
		}
	}
}

func (s *Supervisor) reconcile(ctx context.Context, update configUpdate) {
	if ctx.Err() != nil {
		return
	}
	switch update.state {
	case configNotFound:
		s.disable("secret not found")
		return
	case configUnavailable:
		s.disable("secret unavailable")
		return
	case configSourceInitializationFailed:
		s.logger.Warn("telemetry export configuration source initialization failed", zap.Error(update.err))
		return
	case configLookupFailed:
		s.logger.Warn("telemetry export configuration lookup failed", zap.Error(update.err))
		return
	case configAvailable:
	default:
		return
	}
	if update.value == "" {
		s.disable("secret is empty")
		return
	}
	value := update.value
	if value == s.active {
		s.rejectedConfig = ""
		return
	}
	if value == s.rejectedConfig {
		return
	}
	if value == s.restartConfig {
		return
	}
	cfg, err := parseSecret(value)
	if err != nil {
		s.rejectedConfig = value
		s.logger.Warn("telemetry export configuration is invalid")
		return
	}
	if err := s.replaceChildFn(ctx, cfg); err != nil {
		if ctx.Err() != nil {
			return
		}
		s.logger.Warn("telemetry export collector failed to start")
		if s.active != "" {
			s.rejectedConfig = value
			if previous, parseErr := parseSecret(s.active); parseErr == nil {
				if rollbackErr := s.replaceChildFn(ctx, previous); rollbackErr != nil {
					s.scheduleRestart(s.active)
				} else {
					s.restartConfig = ""
					s.nextStart = time.Time{}
				}
			}
		} else {
			s.scheduleRestart(value)
		}
		return
	}
	s.active = value
	s.rejectedConfig = ""
	s.restartConfig = ""
	s.backoff = time.Second
	s.nextStart = time.Time{}
	s.logEnabled(cfg)
}

func (s *Supervisor) disable(reason string) {
	s.stopChildFn()
	s.active = ""
	s.rejectedConfig = ""
	s.restartConfig = ""
	s.nextStart = time.Time{}
	if !s.reported || s.collectorEnabled {
		s.logger.Info("runner telemetry export collector disabled",
			zap.Bool("telemetry_export.enabled", false),
			zap.Bool("audit_export.enabled", false),
			zap.String("telemetry_export.reason", reason),
		)
	}
	s.reported = true
	s.collectorEnabled = false
}

func (s *Supervisor) logEnabled(cfg config) {
	fields := []zap.Field{
		zap.Bool("telemetry_export.enabled", true),
		zap.Bool("audit_export.enabled", cfg.AuditLogsEnabled),
	}
	if cfg.AuditLogsEnabled {
		endpoint, _ := url.Parse(cfg.OTLPHTTP.Endpoint)
		fields = append(fields,
			zap.String("audit_export.exporter", "otlphttp"),
			zap.String("audit_export.backend", endpoint.Host),
		)
	}
	s.logger.Info("runner telemetry export collector enabled", fields...)
	s.reported = true
	s.collectorEnabled = true
}

func (s *Supervisor) replaceChild(ctx context.Context, cfg config) error {
	contents, environment, err := collectorConfig(cfg)
	if err != nil {
		return err
	}
	s.mu.Lock()
	previous := s.child
	s.mu.Unlock()
	child, err := telemetryexport.StartCollector(ctx, previous, telemetryexport.CollectorOptions{
		Binary: collectorBinary, Config: contents, HealthURL: collectorHealthURL, Environment: environment,
		Args: []string{"--feature-gates=service.AllowNoPipelines"}, Logger: s.logger.Named("telemetry-export-collector"),
	})
	s.mu.Lock()
	s.child = child
	s.mu.Unlock()
	return err
}

func (s *Supervisor) stopChild() {
	s.mu.Lock()
	child := s.child
	s.child = nil
	s.mu.Unlock()
	child.Stop()
}

func (s *Supervisor) restartCrashed(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	s.mu.Lock()
	child := s.child
	s.mu.Unlock()
	if child != nil {
		select {
		case <-child.Done():
			s.stopChild()
			s.logger.Warn("telemetry export collector exited; scheduling restart")
			if time.Since(child.StartedAt()) >= 30*time.Second {
				s.backoff = time.Second
			}
			s.scheduleRestart(s.active)
		default:
		}
	}
	if s.restartConfig == "" || s.nextStart.IsZero() || time.Now().Before(s.nextStart) {
		return
	}
	cfg, err := parseSecret(s.restartConfig)
	if err != nil {
		return
	}
	if s.backoff < 30*time.Second {
		s.backoff *= 2
		if s.backoff > 30*time.Second {
			s.backoff = 30 * time.Second
		}
	}
	s.nextStart = time.Now().Add(s.backoff)
	if err := s.replaceChildFn(ctx, cfg); err != nil {
		s.logger.Warn("telemetry export collector restart failed")
		return
	}
	s.active = s.restartConfig
	s.restartConfig = ""
	s.nextStart = time.Time{}
	s.logEnabled(cfg)
}

func (s *Supervisor) scheduleRestart(value string) {
	s.restartConfig = value
	s.nextStart = time.Now().Add(s.backoff)
}
