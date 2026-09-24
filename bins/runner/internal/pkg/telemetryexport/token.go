package telemetryexport

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	telemetryTokenRequestTimeout = 15 * time.Second
	telemetryTokenMaxLifetime    = 10 * time.Minute
	telemetryTokenMaxSize        = 16 * 1024
	tokenRetryInitial            = time.Second
	tokenRetryMax                = 30 * time.Second
)

type telemetryAccessTokenClient interface {
	CreateTelemetryAccessToken(context.Context) (*models.ServiceCreateTelemetryAccessTokenResponse, error)
}

type tokenLifecycle interface {
	Enable(context.Context) error
	Disable()
}

type tokenManager struct {
	client       telemetryAccessTokenClient
	logger       *zap.Logger
	directory    string
	path         string
	renewalDelay func(time.Duration) time.Duration
	retryInitial time.Duration
	retryMax     time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func newTokenManager(client telemetryAccessTokenClient, logger *zap.Logger) *tokenManager {
	return &tokenManager{
		client:       client,
		logger:       logger,
		directory:    vendorTokenDir,
		path:         vendorTokenPath,
		renewalDelay: randomizedRenewalDelay,
		retryInitial: tokenRetryInitial,
		retryMax:     tokenRetryMax,
	}
}

func (m *tokenManager) Enable(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		return nil
	}

	lifetime, err := m.issue(ctx)
	if err != nil {
		_ = os.RemoveAll(m.directory)
		return err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	m.cancel = cancel
	m.done = done
	go m.run(runCtx, done, lifetime)
	m.logger.Info("telemetry access token issued")
	return nil
}

func (m *tokenManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		<-m.done
		m.cancel = nil
		m.done = nil
	}
	if err := os.RemoveAll(m.directory); err != nil {
		m.logger.Warn("unable to remove telemetry access token directory", zap.Error(err))
	}
}

func (m *tokenManager) run(ctx context.Context, done chan<- struct{}, lifetime time.Duration) {
	defer close(done)
	delay := m.renewalDelay(lifetime)
	backoff := m.retryInitial
	for {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}

		renewCtx, cancel := context.WithTimeout(ctx, telemetryTokenRequestTimeout)
		newLifetime, err := m.issue(renewCtx)
		cancel()
		if err != nil {
			m.logger.Warn("telemetry access token renewal failed", zap.Error(err))
			delay = backoff
			backoff *= 2
			if backoff > m.retryMax {
				backoff = m.retryMax
			}
			continue
		}

		m.logger.Info("telemetry access token renewed")
		lifetime = newLifetime
		delay = m.renewalDelay(lifetime)
		backoff = m.retryInitial
	}
}

func (m *tokenManager) issue(ctx context.Context) (time.Duration, error) {
	requestCtx, cancel := context.WithTimeout(ctx, telemetryTokenRequestTimeout)
	defer cancel()

	response, err := m.client.CreateTelemetryAccessToken(requestCtx)
	if err != nil {
		return 0, fmt.Errorf("create telemetry access token: %w", err)
	}
	if response == nil || response.AccessToken == "" || len(response.AccessToken) > telemetryTokenMaxSize || strings.ContainsAny(response.AccessToken, " \t\r\n\x00") {
		return 0, fmt.Errorf("telemetry access token response contains an invalid access token")
	}
	if response.TokenType != "Bearer" {
		return 0, fmt.Errorf("telemetry access token response contains an invalid token type")
	}
	if response.ExpiresIn <= 0 || response.ExpiresIn > int64(telemetryTokenMaxLifetime/time.Second) {
		return 0, fmt.Errorf("telemetry access token response contains an invalid lifetime")
	}
	lifetime := time.Duration(response.ExpiresIn) * time.Second
	if err := writeTokenAtomically(m.directory, m.path, response.AccessToken); err != nil {
		return 0, fmt.Errorf("store telemetry access token: %w", err)
	}
	return lifetime, nil
}

func randomizedRenewalDelay(lifetime time.Duration) time.Duration {
	base := lifetime * 6 / 10
	window := lifetime / 10
	if window <= 0 {
		return base
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(window)))
	if err != nil {
		return base
	}
	return base + time.Duration(n.Int64())
}

func writeTokenAtomically(directory, path, token string) error {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return fmt.Errorf("set token directory permissions: %w", err)
	}

	f, err := os.CreateTemp(directory, "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary token file: %w", err)
	}
	temporaryPath := f.Name()
	defer os.Remove(temporaryPath)
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return fmt.Errorf("set temporary token file permissions: %w", err)
	}
	if _, err := f.WriteString(token); err != nil {
		f.Close()
		return fmt.Errorf("write temporary token file: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync temporary token file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temporary token file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace token file: %w", err)
	}

	dir, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open token directory: %w", err)
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync token directory: %w", err)
	}
	return nil
}
