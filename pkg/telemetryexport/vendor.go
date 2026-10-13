package telemetryexport

import (
	"context"
	"maps"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	vendorSettingsRefreshInterval  = 15 * time.Second
	vendorSettingsRequestTimeout   = 15 * time.Second
	vendorCollectorRestartInterval = time.Second
)

// Settings is the remotely selected relay destination and resource metadata.
// Adapters map control-plane responses into Settings at their boundary.
type Settings struct {
	Enabled            bool
	RelayEndpoint      string
	ResourceAttributes map[string]string
}

// Options separates control-plane settings and host lifecycle integration from
// token renewal, reconciliation, and Collector process supervision. An empty
// InstallID makes the supervisor ineligible: Run removes stale credentials and
// returns without fetching settings. A nil InitialSettings fetches settings
// before the first reconciliation.
type Options struct {
	InstallID       string
	Binary          string
	Config          Config
	Client          telemetryAccessTokenClient
	Logger          *zap.Logger
	InitialSettings *Settings
	FetchSettings   func(context.Context) (Settings, error)
	OnEnabled       func(bool)
}

type VendorSupervisor struct {
	installID           string
	binary              string
	config              Config
	onEnabled           func(bool)
	ready               atomic.Bool
	initialSettings     *Settings
	activeEndpoint      string
	desiredEndpoint     string
	activeAttributes    map[string]string
	desiredAttributes   map[string]string
	rejectedEndpoint    string
	logger              *zap.Logger
	tokens              tokenLifecycle
	mu                  sync.Mutex
	child               *Collector
	backoff             time.Duration
	nextStart           time.Time
	enabled             bool
	disabled            bool
	reportedEnabled     bool
	settingsUnavailable bool

	fetchSettingsFn func(context.Context) (Settings, error)
	replaceChildFn  func(context.Context, string, map[string]string) error
	stopChildFn     func()
}

func NewVendor(options Options) *VendorSupervisor {
	s := &VendorSupervisor{
		installID:       options.InstallID,
		binary:          options.Binary,
		config:          options.Config,
		onEnabled:       options.OnEnabled,
		logger:          options.Logger,
		tokens:          newTokenManager(options.Client, options.InstallID, options.Config.TokenDirectory, options.Logger),
		backoff:         time.Second,
		fetchSettingsFn: options.FetchSettings,
	}
	if options.InitialSettings != nil {
		initial := *options.InitialSettings
		initial.ResourceAttributes = maps.Clone(initial.ResourceAttributes)
		s.initialSettings = &initial
	}
	s.replaceChildFn = s.replaceChild
	s.stopChildFn = s.stopChild
	return s
}

// Run polls settings even while forwarding is disabled. It stops the child
// process and removes credentials before every return; callers own
// cancellation and must join Run. Call once per supervisor.
func (s *VendorSupervisor) Run(ctx context.Context) {
	defer func() {
		s.stopChildFn()
		s.tokens.Disable()
		s.ready.Store(false)
	}()
	if s.installID == "" {
		return
	}

	if s.initialSettings != nil {
		s.reconcile(ctx, *s.initialSettings)
	} else {
		s.refreshSettings(ctx)
	}
	settingsTicker := time.NewTicker(vendorSettingsRefreshInterval)
	restartTicker := time.NewTicker(vendorCollectorRestartInterval)
	defer settingsTicker.Stop()
	defer restartTicker.Stop()
	for {
		select {
		case <-ctx.Done():
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

func (s *VendorSupervisor) reconcile(ctx context.Context, settings Settings) {
	if !settings.Enabled || settings.RelayEndpoint == "" {
		s.disable()
		return
	}
	if err := validateEndpoint(settings.RelayEndpoint); err != nil {
		if settings.RelayEndpoint != s.rejectedEndpoint {
			s.logger.Warn("vendor telemetry relay endpoint is invalid; retaining active configuration", zap.Error(err))
		}
		if s.enabled {
			s.desiredEndpoint = s.activeEndpoint
			s.desiredAttributes = maps.Clone(s.activeAttributes)
			s.nextStart = time.Time{}
			s.backoff = time.Second
		} else {
			s.disable()
		}
		s.rejectedEndpoint = settings.RelayEndpoint
		return
	}
	s.rejectedEndpoint = ""
	if settings.RelayEndpoint == s.desiredEndpoint && maps.Equal(settings.ResourceAttributes, s.desiredAttributes) {
		return
	}

	if s.desiredEndpoint != "" && s.desiredEndpoint != settings.RelayEndpoint {
		s.ready.Store(false)
		s.stopChildFn()
		s.tokens.Disable()
		s.activeEndpoint = ""
		s.activeAttributes = nil
		s.enabled = false
	}
	s.desiredEndpoint = settings.RelayEndpoint
	s.desiredAttributes = maps.Clone(settings.ResourceAttributes)
	s.disabled = false
	s.backoff = time.Second
	s.nextStart = time.Time{}
	if s.enabled && s.activeEndpoint == s.desiredEndpoint && maps.Equal(s.activeAttributes, s.desiredAttributes) {
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
	s.activeAttributes = nil
	s.desiredAttributes = nil
	s.rejectedEndpoint = ""
	s.nextStart = time.Time{}
	s.backoff = time.Second
	s.enabled = false
	s.disabled = true
	s.ready.Store(true)
	s.logger.Info("vendor telemetry export collector disabled")
	s.reportEnabled(false)
}

func (s *VendorSupervisor) startCollector(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	s.ready.Store(false)
	if err := s.tokens.Enable(ctx, s.desiredEndpoint); err != nil {
		s.logger.Warn("vendor telemetry access token unavailable", zap.Error(err))
		s.scheduleRestart()
		return
	}

	previousEndpoint := s.activeEndpoint
	previousAttributes := s.activeAttributes
	previousEnabled := s.enabled
	if err := s.replaceChildFn(ctx, s.desiredEndpoint, s.desiredAttributes); err != nil {
		if ctx.Err() != nil {
			return
		}
		s.logger.Warn("vendor telemetry export collector failed to start", zap.Error(err))
		if previousEnabled && previousEndpoint != "" && (previousEndpoint != s.desiredEndpoint || !maps.Equal(previousAttributes, s.desiredAttributes)) {
			if rollbackErr := s.replaceChildFn(ctx, previousEndpoint, previousAttributes); rollbackErr == nil {
				s.activeEndpoint = previousEndpoint
				s.activeAttributes = previousAttributes
				s.enabled = true
				s.ready.Store(true)
				s.scheduleRestart()
				return
			} else {
				s.logger.Error("vendor telemetry export collector rollback failed", zap.Error(rollbackErr))
			}
		}
		s.tokens.Disable()
		s.activeEndpoint = ""
		s.activeAttributes = nil
		s.enabled = false
		s.ready.Store(false)
		s.scheduleRestart()
		return
	}

	s.enabled = true
	s.activeEndpoint = s.desiredEndpoint
	s.activeAttributes = maps.Clone(s.desiredAttributes)
	s.nextStart = time.Time{}
	s.ready.Store(true)
	endpoint, _ := url.Parse(s.activeEndpoint)
	s.logger.Info("vendor telemetry export collector enabled",
		zap.String("vendor_telemetry_export.backend", endpoint.Host),
		zap.Bool("vendor_telemetry_export.logs", true),
		zap.Bool("vendor_telemetry_export.metrics", true),
		zap.Bool("vendor_telemetry_export.traces", true),
	)
	s.reportEnabled(true)
}

// Ready distinguishes a healthy, intentionally disabled controller from one that
// has not loaded settings or cannot start its enabled Collector.
func (s *VendorSupervisor) Ready() bool { return s.ready.Load() }

func (s *VendorSupervisor) reportEnabled(enabled bool) {
	if s.reportedEnabled == enabled {
		return
	}
	s.reportedEnabled = enabled
	if s.onEnabled != nil {
		s.onEnabled(enabled)
	}
}

func (s *VendorSupervisor) replaceChild(ctx context.Context, endpoint string, attributes map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	contents, err := CollectorConfig(s.config, endpoint, attributes)
	if err != nil {
		return err
	}
	s.mu.Lock()
	previous := s.child
	s.mu.Unlock()
	child, err := StartCollector(ctx, previous, CollectorOptions{
		Binary: s.binary, Config: contents, HealthURL: "http://" + s.config.HealthAddress + "/",
		Logger: s.logger.Named("vendor-telemetry-export-collector"),
	})
	s.mu.Lock()
	s.child = child
	s.mu.Unlock()
	return err
}

func (s *VendorSupervisor) stopChild() {
	s.mu.Lock()
	child := s.child
	s.child = nil
	s.mu.Unlock()
	child.Stop()
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
			s.activeAttributes = nil
			s.ready.Store(false)
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
