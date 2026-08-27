package telemetryexport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	collectorBinary             = "/bin/nuon-runner-otelcol"
	collectorHealthURL          = "http://127.0.0.1:13133/"
	collectorStartTimeout       = 5 * time.Second
	collectorHealthPollInterval = 50 * time.Millisecond
	secretSyncInterval          = 30 * time.Second
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
	child            *childProcess
	active           string
	rejectedConfig   string
	restartConfig    string
	backoff          time.Duration
	nextStart        time.Time
	reported         bool
	collectorEnabled bool

	replaceChildFn func(config) error
	stopChildFn    func()
}

type childProcess struct {
	cmd           *exec.Cmd
	tempDir       string
	done          chan struct{}
	startedAt     time.Time
	outputWriters []io.Closer
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
			s.stopChildFn()
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
	updates := source.Watch(ctx, secretSyncInterval)
	crash := time.NewTicker(time.Second)
	defer crash.Stop()
	for {
		select {
		case <-ctx.Done():
			s.stopChildFn()
			return
		case update, ok := <-updates:
			if !ok {
				s.stopChildFn()
				return
			}
			s.reconcile(update)
		case <-crash.C:
			s.restartCrashed(ctx)
		}
	}
}

func (s *Supervisor) reconcile(update configUpdate) {
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
	if err := s.replaceChildFn(cfg); err != nil {
		s.logger.Warn("telemetry export collector failed to start")
		if s.active != "" {
			s.rejectedConfig = value
			if previous, parseErr := parseSecret(s.active); parseErr == nil {
				if rollbackErr := s.replaceChildFn(previous); rollbackErr != nil {
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

func (s *Supervisor) replaceChild(cfg config) error {
	if _, err := os.Stat(collectorBinary); err != nil {
		return err
	}
	contents, environment, err := collectorConfig(cfg)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp("", "nuon-telemetry-export-")
	if err != nil {
		return err
	}
	path := filepath.Join(tempDir, "collector.yaml")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		os.RemoveAll(tempDir)
		return err
	}
	cmd := exec.Command(collectorBinary, "--config", path, "--feature-gates=service.AllowNoPipelines")
	cmd.Env = childEnvironment(environment)
	stdout := &zapio.Writer{Log: s.logger.Named("telemetry-export-collector").With(zap.String("stream", "stdout")), Level: zapcore.WarnLevel}
	stderr := &zapio.Writer{Log: s.logger.Named("telemetry-export-collector").With(zap.String("stream", "stderr")), Level: zapcore.WarnLevel}
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
	if err := waitForCollector(child, collectorHealthURL); err != nil {
		s.stopChild()
		return err
	}
	return nil
}

func waitForCollector(child *childProcess, healthURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), collectorStartTimeout)
	defer cancel()
	client := &http.Client{Timeout: collectorHealthPollInterval}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				return nil
			}
		}
		select {
		case <-child.done:
			return fmt.Errorf("telemetry export collector exited before becoming healthy")
		case <-ctx.Done():
			return fmt.Errorf("telemetry export collector did not become healthy: %w", ctx.Err())
		case <-time.After(collectorHealthPollInterval):
		}
	}
}

func (s *Supervisor) stopChild() {
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

func (s *Supervisor) restartCrashed(_ context.Context) {
	s.mu.Lock()
	child := s.child
	s.mu.Unlock()
	if child != nil {
		select {
		case <-child.done:
			s.stopChild()
			s.logger.Warn("telemetry export collector exited; scheduling restart")
			if time.Since(child.startedAt) >= 30*time.Second {
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
	if err := s.replaceChildFn(cfg); err != nil {
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

func childEnvironment(secretHeaders []string) []string {
	allowed := map[string]struct{}{
		"SSL_CERT_FILE": {}, "SSL_CERT_DIR": {},
		"HTTP_PROXY": {}, "HTTPS_PROXY": {}, "NO_PROXY": {},
		"http_proxy": {}, "https_proxy": {}, "no_proxy": {},
	}
	environment := make([]string, 0, len(secretHeaders)+len(allowed))
	for _, value := range os.Environ() {
		name, _, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		if _, ok := allowed[name]; ok {
			environment = append(environment, value)
		}
	}
	return append(environment, secretHeaders...)
}
