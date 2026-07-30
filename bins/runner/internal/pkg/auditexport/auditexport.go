package auditexport

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

const collectorBinary = "/bin/nuon-runner-otelcol"
const secretSyncInterval = 30 * time.Second

var Module = fx.Options(fx.Provide(newAWSFactory, New), fx.Invoke(func(*Supervisor) {}))

type secretGetter interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type clientFactory interface {
	New(context.Context) (secretGetter, error)
}

type awsFactory struct{}

func newAWSFactory() clientFactory { return awsFactory{} }

func (awsFactory) New(ctx context.Context) (secretGetter, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Region == "" {
		region, regionErr := imds.NewFromConfig(cfg).GetRegion(ctx, nil)
		if regionErr != nil {
			return nil, regionErr
		}
		cfg.Region = region.Region
	}
	return secretsmanager.NewFromConfig(cfg), nil
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Settings  *settings.Settings
	Logger    *zap.Logger `name:"system"`
	Factory   clientFactory
}

type Supervisor struct {
	installID string
	platform  string
	local     bool
	logger    *zap.Logger
	factory   clientFactory
	cancel    context.CancelFunc
	done      chan struct{}
	mu        sync.Mutex
	child     *childProcess
	active    string
	backoff   time.Duration
	nextStart time.Time
	reported  bool
	enabled   bool

	replaceChildFn func(secretConfig) error
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
	s := &Supervisor{installID: params.Settings.Metadata["install.id"], platform: params.Settings.Platform, local: params.Settings.Cfg.IsNuonctl, logger: params.Logger, factory: params.Factory, done: make(chan struct{}), backoff: time.Second}
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
	if s.local || s.installID == "" || !strings.HasPrefix(strings.ToLower(s.platform), "aws") {
		return
	}
	secretSync := time.NewTicker(secretSyncInterval)
	defer secretSync.Stop()
	crash := time.NewTicker(time.Second)
	defer crash.Stop()
	var client secretGetter
	refresh := func() {
		if client == nil {
			var err error
			client, err = s.factory.New(ctx)
			if err != nil {
				client = nil
				s.logger.Warn("audit export AWS initialization failed")
				return
			}
		}
		s.reconcile(ctx, client)
	}
	refresh()
	for {
		select {
		case <-ctx.Done():
			s.stopChildFn()
			return
		case <-secretSync.C:
			refresh()
		case <-crash.C:
			s.restartCrashed(ctx)
		}
	}
}

func (s *Supervisor) reconcile(ctx context.Context, client secretGetter) {
	name := "nuon/" + s.installID + "/runner-audit-export"
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &name})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			s.disable("secret not found")
		} else {
			s.logger.Warn("audit export secret lookup failed")
		}
		return
	}
	if result.SecretString == nil || *result.SecretString == "" {
		s.disable("secret is empty")
		return
	}
	value := *result.SecretString
	if value == s.active {
		return
	}
	cfg, err := parseSecret(value)
	if err != nil {
		s.logger.Warn("audit export configuration is invalid")
		return
	}
	if err := s.replaceChildFn(cfg); err != nil {
		s.logger.Warn("audit export collector failed to start")
		if s.active != "" {
			if previous, parseErr := parseSecret(s.active); parseErr == nil {
				if rollbackErr := s.replaceChildFn(previous); rollbackErr != nil {
					s.nextStart = time.Now().Add(s.backoff)
				}
			}
		}
		return
	}
	s.active = value
	s.backoff = time.Second
	s.nextStart = time.Time{}
	s.logEnabled(cfg)
}

func (s *Supervisor) disable(reason string) {
	s.stopChildFn()
	s.active = ""
	s.nextStart = time.Time{}
	if !s.reported || s.enabled {
		s.logger.Info("runner audit export disabled",
			zap.Bool("audit_export.enabled", false),
			zap.String("audit_export.reason", reason),
		)
	}
	s.reported = true
	s.enabled = false
}

func (s *Supervisor) logEnabled(cfg secretConfig) {
	endpoint, _ := url.Parse(cfg.Exporters.OTLPHTTP.Endpoint)
	s.logger.Info("runner audit export enabled",
		zap.Bool("audit_export.enabled", true),
		zap.String("audit_export.exporter", "otlphttp"),
		zap.String("audit_export.backend", endpoint.Host),
	)
	s.reported = true
	s.enabled = true
}

func (s *Supervisor) replaceChild(cfg secretConfig) error {
	if _, err := os.Stat(collectorBinary); err != nil {
		return err
	}
	contents, environment, err := collectorConfig(cfg)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp("", "nuon-audit-export-")
	if err != nil {
		return err
	}
	path := filepath.Join(tempDir, "collector.yaml")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		os.RemoveAll(tempDir)
		return err
	}
	cmd := exec.Command(collectorBinary, "--config", path)
	cmd.Env = childEnvironment(environment)
	stdout := &zapio.Writer{Log: s.logger.Named("audit-export-collector").With(zap.String("stream", "stdout")), Level: zapcore.WarnLevel}
	stderr := &zapio.Writer{Log: s.logger.Named("audit-export-collector").With(zap.String("stream", "stderr")), Level: zapcore.WarnLevel}
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
	return nil
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
			s.logger.Warn("audit export collector exited; scheduling restart")
			if time.Since(child.startedAt) >= 30*time.Second {
				s.backoff = time.Second
			}
			s.nextStart = time.Now().Add(s.backoff)
		default:
		}
	}
	if s.active == "" || s.nextStart.IsZero() || time.Now().Before(s.nextStart) {
		return
	}
	cfg, err := parseSecret(s.active)
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
		s.logger.Warn("audit export collector restart failed")
		return
	}
	s.nextStart = time.Time{}
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
