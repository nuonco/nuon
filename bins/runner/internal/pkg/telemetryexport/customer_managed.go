package telemetryexport

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

const CustomerManagedOTLPLogsEndpoint = "http://127.0.0.1:4318/v1/logs"

var CustomerManagedModule = fx.Options(
	fx.Provide(NewCustomerManaged),
	fx.Invoke(func(*CustomerManagedSupervisor) {}),
)

type CustomerManagedParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Settings  *settings.Settings
	Logger    *zap.Logger `name:"system"`
}

type CustomerManagedSupervisor struct {
	logDir string
	logger *zap.Logger

	mu        sync.Mutex
	child     *customerManagedCollectorProcess
	cancel    context.CancelFunc
	done      chan struct{}
	configDir string
}

type customerManagedCollectorProcess struct {
	cmd     *exec.Cmd
	done    chan struct{}
	writers []io.Closer
}

func NewCustomerManaged(params CustomerManagedParams) *CustomerManagedSupervisor {
	s := &CustomerManagedSupervisor{
		logDir: params.Settings.Cfg.OTELLogDir,
		logger: params.Logger,
		done:   make(chan struct{}),
	}
	params.Lifecycle.Append(fx.Hook{OnStart: s.start, OnStop: s.stop})
	return s
}

func (s *CustomerManagedSupervisor) start(ctx context.Context) error {
	if s.logDir == "" {
		return fmt.Errorf("offline OTEL log directory is required")
	}
	if err := os.MkdirAll(s.logDir, 0o700); err != nil {
		return fmt.Errorf("create offline OTEL log directory: %w", err)
	}
	contents, err := customerManagedCollectorConfig(s.logDir)
	if err != nil {
		return err
	}
	s.configDir, err = os.MkdirTemp("", "nuon-customer-managed-otel-")
	if err != nil {
		return fmt.Errorf("create offline OTEL config directory: %w", err)
	}
	configPath := filepath.Join(s.configDir, "collector.yaml")
	if err := os.WriteFile(configPath, contents, 0o600); err != nil {
		_ = os.RemoveAll(s.configDir)
		return fmt.Errorf("write offline OTEL config: %w", err)
	}
	if err := s.launch(configPath); err != nil {
		_ = os.RemoveAll(s.configDir)
		return err
	}
	if err := s.waitUntilReady(ctx); err != nil {
		s.stopChild()
		_ = os.RemoveAll(s.configDir)
		return err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.supervise(runCtx, configPath)
	s.logger.Info("offline OTEL collector enabled", zap.String("otel.logs.path", filepath.Join(s.logDir, "otel.jsonl")))
	return nil
}

func (s *CustomerManagedSupervisor) stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	s.stopChild()
	select {
	case <-s.done:
	case <-ctx.Done():
		return ctx.Err()
	}
	if s.configDir != "" {
		_ = os.RemoveAll(s.configDir)
	}
	return nil
}

func (s *CustomerManagedSupervisor) supervise(ctx context.Context, configPath string) {
	defer close(s.done)
	backoff := time.Second
	for {
		child := s.currentChild()
		if child == nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-child.done:
		}
		s.clearChild(child)
		s.logger.Warn("offline OTEL collector exited; restarting", zap.Duration("backoff", backoff))
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if ctx.Err() != nil {
			return
		}
		if err := s.launch(configPath); err != nil {
			s.logger.Error("restart offline OTEL collector", zap.Error(err))
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		if ctx.Err() != nil {
			s.stopChild()
			return
		}
		backoff = time.Second
	}
}

func (s *CustomerManagedSupervisor) launch(configPath string) error {
	path := os.Getenv("NUON_RUNNER_OTELCOL_PATH")
	if path == "" {
		path = collectorBinary
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("offline OTEL collector binary %s: %w", path, err)
	}
	cmd := exec.Command(path, "--config", configPath)
	cmd.Env = childEnvironment(nil)
	stdout := &zapio.Writer{Log: s.logger.Named("customer-managed-otel-collector").With(zap.String("stream", "stdout")), Level: zapcore.WarnLevel}
	stderr := &zapio.Writer{Log: s.logger.Named("customer-managed-otel-collector").With(zap.String("stream", "stderr")), Level: zapcore.WarnLevel}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		return fmt.Errorf("start offline OTEL collector: %w", err)
	}
	child := &customerManagedCollectorProcess{cmd: cmd, done: make(chan struct{}), writers: []io.Closer{stdout, stderr}}
	s.mu.Lock()
	s.child = child
	s.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		for _, writer := range child.writers {
			_ = writer.Close()
		}
		close(child.done)
	}()
	return nil
}

func (s *CustomerManagedSupervisor) waitUntilReady(ctx context.Context) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	client := &http.Client{Timeout: time.Second}
	for {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:13133/", nil)
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		child := s.currentChild()
		if child != nil {
			select {
			case <-child.done:
				return fmt.Errorf("offline OTEL collector exited before becoming ready")
			default:
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("offline OTEL collector did not become ready")
		case <-ticker.C:
		}
	}
}

func (s *CustomerManagedSupervisor) stopChild() {
	s.mu.Lock()
	child := s.child
	s.child = nil
	s.mu.Unlock()
	if child == nil || child.cmd.Process == nil {
		return
	}
	_ = child.cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-child.done:
	case <-time.After(5 * time.Second):
		_ = child.cmd.Process.Kill()
		<-child.done
	}
}

func (s *CustomerManagedSupervisor) currentChild() *customerManagedCollectorProcess {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.child
}

func (s *CustomerManagedSupervisor) clearChild(child *customerManagedCollectorProcess) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.child == child {
		s.child = nil
	}
}

func customerManagedCollectorConfig(logDir string) ([]byte, error) {
	document := map[string]any{
		"extensions": map[string]any{"health_check": map[string]any{"endpoint": "127.0.0.1:13133"}},
		"receivers":  map[string]any{"otlp": map[string]any{"protocols": map[string]any{"http": map[string]any{"endpoint": "127.0.0.1:4318"}}}},
		"processors": map[string]any{
			"memory_limiter": map[string]any{"check_interval": "1s", "limit_mib": 128, "spike_limit_mib": 32},
			"batch":          map[string]any{"send_batch_size": 512, "timeout": "1s"},
		},
		"exporters": map[string]any{
			"file/logs": map[string]any{
				"path": filepath.Join(logDir, "otel.jsonl"), "format": "json",
				"rotation": map[string]any{"max_megabytes": 10, "max_days": 0, "max_backups": 10, "localtime": false},
			},
		},
		"service": map[string]any{
			"extensions": []string{"health_check"},
			"pipelines": map[string]any{
				"logs": map[string]any{"receivers": []string{"otlp"}, "processors": []string{"memory_limiter", "batch"}, "exporters": []string{"file/logs"}},
			},
			"telemetry": map[string]any{"logs": map[string]any{"level": "warn"}},
		},
	}
	contents, err := yaml.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("encode offline OTEL collector config: %w", err)
	}
	return contents, nil
}
