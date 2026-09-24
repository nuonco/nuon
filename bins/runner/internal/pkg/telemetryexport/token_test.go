package telemetryexport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type tokenClientResult struct {
	response *models.ServiceCreateTelemetryAccessTokenResponse
	err      error
}

type fakeTokenClient struct {
	mu      sync.Mutex
	results []tokenClientResult
	calls   chan struct{}
}

func (c *fakeTokenClient) CreateTelemetryAccessToken(context.Context) (*models.ServiceCreateTelemetryAccessTokenResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case c.calls <- struct{}{}:
	default:
	}
	if len(c.results) == 0 {
		return nil, errors.New("no token response configured")
	}
	result := c.results[0]
	c.results = c.results[1:]
	return result.response, result.err
}

func validTokenResponse(token string) *models.ServiceCreateTelemetryAccessTokenResponse {
	return &models.ServiceCreateTelemetryAccessTokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: 600}
}

func testTokenManager(t *testing.T, client *fakeTokenClient) *tokenManager {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "auth")
	return &tokenManager{
		client:       client,
		logger:       zap.NewNop(),
		directory:    directory,
		path:         filepath.Join(directory, "access-token"),
		renewalDelay: func(time.Duration) time.Duration { return time.Hour },
		retryInitial: time.Hour,
		retryMax:     time.Hour,
	}
}

func waitForTokenCall(t *testing.T, calls <-chan struct{}) {
	t.Helper()
	select {
	case <-calls:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for telemetry token request")
	}
}

func TestTokenManagerWritesRenewsAndRemovesProtectedToken(t *testing.T) {
	client := &fakeTokenClient{
		results: []tokenClientResult{{response: validTokenResponse("initial.jwt")}, {response: validTokenResponse("renewed.jwt")}},
		calls:   make(chan struct{}, 2),
	}
	manager := testTokenManager(t, client)
	renewals := 0
	manager.renewalDelay = func(time.Duration) time.Duration {
		renewals++
		if renewals == 1 {
			return time.Millisecond
		}
		return time.Hour
	}
	defer manager.Disable()

	if err := manager.Enable(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitForTokenCall(t, client.calls)
	waitForTokenCall(t, client.calls)

	var contents []byte
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		contents, _ = os.ReadFile(manager.path)
		if string(contents) == "renewed.jwt" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if string(contents) != "renewed.jwt" {
		t.Fatalf("token file was not atomically renewed: %q", contents)
	}
	directoryInfo, err := os.Stat(manager.directory)
	if err != nil {
		t.Fatal(err)
	}
	fileInfo, err := os.Stat(manager.path)
	if err != nil {
		t.Fatal(err)
	}
	if directoryInfo.Mode().Perm() != 0o700 || fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected token permissions: directory=%o file=%o", directoryInfo.Mode().Perm(), fileInfo.Mode().Perm())
	}

	manager.Disable()
	if _, err := os.Stat(manager.directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("token directory still exists after shutdown: %v", err)
	}
}

func TestTokenManagerRetainsCurrentTokenAfterRenewalFailure(t *testing.T) {
	client := &fakeTokenClient{
		results: []tokenClientResult{{response: validTokenResponse("current.jwt")}, {err: errors.New("ctl-api unavailable")}},
		calls:   make(chan struct{}, 2),
	}
	manager := testTokenManager(t, client)
	manager.renewalDelay = func(time.Duration) time.Duration { return time.Millisecond }
	defer manager.Disable()

	if err := manager.Enable(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitForTokenCall(t, client.calls)
	waitForTokenCall(t, client.calls)
	contents, err := os.ReadFile(manager.path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "current.jwt" {
		t.Fatalf("failed renewal replaced the current token: %q", contents)
	}
}

func TestTokenManagerRejectsInvalidResponsesWithoutLeavingCredentials(t *testing.T) {
	tests := map[string]*models.ServiceCreateTelemetryAccessTokenResponse{
		"nil":               nil,
		"empty token":       {TokenType: "Bearer", ExpiresIn: 600},
		"token whitespace":  {AccessToken: "invalid token", TokenType: "Bearer", ExpiresIn: 600},
		"wrong type":        {AccessToken: "token.jwt", TokenType: "bearer", ExpiresIn: 600},
		"zero lifetime":     {AccessToken: "token.jwt", TokenType: "Bearer"},
		"long lifetime":     {AccessToken: "token.jwt", TokenType: "Bearer", ExpiresIn: 601},
		"overflow lifetime": {AccessToken: "token.jwt", TokenType: "Bearer", ExpiresIn: 1<<63 - 1},
	}
	for name, response := range tests {
		t.Run(name, func(t *testing.T) {
			client := &fakeTokenClient{results: []tokenClientResult{{response: response}}, calls: make(chan struct{}, 1)}
			manager := testTokenManager(t, client)
			if err := manager.Enable(context.Background()); err == nil {
				t.Fatal("expected invalid token response to fail")
			}
			if _, err := os.Stat(manager.directory); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("invalid response left token state on disk: %v", err)
			}
		})
	}
}

func TestRandomizedRenewalDelayStaysWithinWindow(t *testing.T) {
	lifetime := 10 * time.Minute
	for range 100 {
		delay := randomizedRenewalDelay(lifetime)
		if delay < 6*time.Minute || delay >= 7*time.Minute {
			t.Fatalf("renewal delay %s is outside the 60-70%% lifetime window", delay)
		}
	}
}
