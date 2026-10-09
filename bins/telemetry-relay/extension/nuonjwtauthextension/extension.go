package nuonjwtauthextension

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensionauth"
	"go.uber.org/zap"
)

const (
	tokenType          = "at+jwt"
	tokenScope         = "telemetry:write"
	legacyAudience     = "urn:nuon:telemetry"
	tokenLifetime      = 10 * time.Minute
	tokenClockSkew     = 30 * time.Second
	refreshInterval    = 5 * time.Minute
	maxAccessTokenSize = 16 * 1024
)

var (
	errAuthenticationFailed = errors.New("telemetry authentication failed")
	shortIDPattern          = regexp.MustCompile(`^[0-9a-z]{26}$`)
)

type telemetryClaims struct {
	ClientID    string `json:"client_id"`
	Scope       string `json:"scope"`
	OrgID       string `json:"nuon_org_id"`
	AppID       string `json:"nuon_app_id"`
	InstallID   string `json:"nuon_install_id"`
	RunnerID    string `json:"nuon_runner_id,omitempty"`
	CollectorID string `json:"nuon_collector_id,omitempty"`
	jwt.RegisteredClaims
}

type telemetryJWTAuthExtension struct {
	config Config
	keys   *keyCache
	logger *zap.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var (
	_ extension.Extension  = (*telemetryJWTAuthExtension)(nil)
	_ extensionauth.Server = (*telemetryJWTAuthExtension)(nil)
)

func newExtension(cfg Config, logger *zap.Logger) *telemetryJWTAuthExtension {
	return &telemetryJWTAuthExtension{
		config: cfg,
		keys:   newKeyCache(cfg.JWKSURL),
		logger: logger,
	}
}

func (e *telemetryJWTAuthExtension) Start(ctx context.Context, _ component.Host) error {
	if err := e.keys.refresh(ctx, false); err != nil {
		return fmt.Errorf("load telemetry JWKS: %w", err)
	}

	refreshCtx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(refreshCtx, jwksHTTPTimeout)
				err := e.keys.refresh(ctx, false)
				cancel()
				if err != nil && e.logger != nil {
					e.logger.Warn("unable to refresh telemetry JWKS", zap.Error(err))
				}
			}
		}
	}()
	return nil
}

func (e *telemetryJWTAuthExtension) Shutdown(context.Context) error {
	if e.cancel != nil {
		e.cancel()
	}
	e.wg.Wait()
	e.keys.close()
	return nil
}

func (e *telemetryJWTAuthExtension) Authenticate(ctx context.Context, sources map[string][]string) (context.Context, error) {
	raw, err := bearerToken(sources)
	if err != nil {
		return ctx, errAuthenticationFailed
	}

	principal, err := e.verify(ctx, raw)
	if err != nil {
		return ctx, errAuthenticationFailed
	}

	info := client.FromContext(ctx)
	info.Auth = NewAuthData(principal)
	return client.NewContext(ctx, info), nil
}

func bearerToken(sources map[string][]string) (string, error) {
	var values []string
	for name, sourceValues := range sources {
		if strings.EqualFold(name, "authorization") {
			values = append(values, sourceValues...)
		}
	}
	if len(values) != 1 {
		return "", errAuthenticationFailed
	}

	parts := strings.Fields(values[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" || len(parts[1]) > maxAccessTokenSize || strings.Count(parts[1], ".") != 2 {
		return "", errAuthenticationFailed
	}
	return parts[1], nil
}

func (e *telemetryJWTAuthExtension) verify(ctx context.Context, raw string) (Principal, error) {
	audiences := []string{e.config.Audience}
	if e.config.AllowLegacyAudience {
		audiences = append(audiences, legacyAudience)
	}
	claims := &telemetryClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodRS256 || token.Header["typ"] != tokenType {
			return nil, errAuthenticationFailed
		}
		keyID, ok := token.Header["kid"].(string)
		if !ok || keyID == "" {
			return nil, errAuthenticationFailed
		}
		return e.keys.key(ctx, keyID)
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(e.config.Issuer),
		jwt.WithAudience(audiences...),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(tokenClockSkew),
	)
	if err != nil || !token.Valid {
		return Principal{}, errAuthenticationFailed
	}

	principal := Principal{
		OrgID:       claims.OrgID,
		AppID:       claims.AppID,
		InstallID:   claims.InstallID,
		RunnerID:    claims.RunnerID,
		CollectorID: claims.CollectorID,
	}
	if err := validateClaims(claims, principal); err != nil {
		return Principal{}, errAuthenticationFailed
	}
	return principal, nil
}

func validateClaims(claims *telemetryClaims, principal Principal) error {
	if claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil || claims.ID == "" {
		return errAuthenticationFailed
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt.Time) > tokenLifetime || !claims.ExpiresAt.After(claims.IssuedAt.Time) || claims.NotBefore.After(claims.ExpiresAt.Time) {
		return errAuthenticationFailed
	}
	if len(claims.Audience) != 1 || claims.Scope != tokenScope {
		return errAuthenticationFailed
	}
	if !validPrincipal(principal) {
		return errAuthenticationFailed
	}
	kind, clientID := "runner", principal.RunnerID
	if principal.CollectorID != "" {
		if claims.Audience[0] == legacyAudience {
			return errAuthenticationFailed
		}
		kind, clientID = "collector", principal.CollectorID
	}
	expectedSubject := fmt.Sprintf("org:%s:install:%s:%s:%s", principal.OrgID, principal.InstallID, kind, clientID)
	if claims.ClientID != clientID || claims.Subject != expectedSubject {
		return errAuthenticationFailed
	}
	return nil
}

func validPrincipal(principal Principal) bool {
	if !validID(principal.OrgID, "org") || !validID(principal.AppID, "app") || !validID(principal.InstallID, "inl") {
		return false
	}
	if principal.CollectorID != "" {
		return principal.RunnerID == "" && validID(principal.CollectorID, "acc")
	}
	return validID(principal.RunnerID, "run")
}

func validID(value, prefix string) bool {
	return strings.HasPrefix(value, prefix) && shortIDPattern.MatchString(value)
}
