package nuonjwtauthextension

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.uber.org/zap"
)

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func testJWKS(t *testing.T, keys map[string]*rsa.PrivateKey) []byte {
	t.Helper()
	set := jsonWebKeySet{Keys: make([]jsonWebKey, 0, len(keys))}
	for keyID, key := range keys {
		set.Keys = append(set.Keys, jsonWebKey{
			KeyType:   "RSA",
			KeyID:     keyID,
			Use:       "sig",
			Algorithm: jwt.SigningMethodRS256.Alg(),
			Modulus:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			Exponent:  base64.RawURLEncoding.EncodeToString(bigEndianBytes(key.E)),
		})
	}
	contents, err := json.Marshal(set)
	require.NoError(t, err)
	return contents
}

func bigEndianBytes(value int) []byte {
	var bytes []byte
	for value > 0 {
		bytes = append([]byte{byte(value)}, bytes...)
		value >>= 8
	}
	return bytes
}

func testPrincipal() Principal {
	return Principal{
		OrgID:     "org" + strings.Repeat("a", 23),
		AppID:     "app" + strings.Repeat("b", 23),
		InstallID: "inl" + strings.Repeat("c", 23),
		RunnerID:  "run" + strings.Repeat("d", 23),
	}
}

func testClaims(now time.Time, principal Principal) *telemetryClaims {
	return &telemetryClaims{
		ClientID:  principal.RunnerID,
		Scope:     tokenScope,
		OrgID:     principal.OrgID,
		AppID:     principal.AppID,
		InstallID: principal.InstallID,
		RunnerID:  principal.RunnerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://ctl.example.com",
			Subject:   "org:" + principal.OrgID + ":install:" + principal.InstallID + ":runner:" + principal.RunnerID,
			Audience:  jwt.ClaimStrings{defaultAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        "test-token-id",
		},
	}
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, keyID string, claims *telemetryClaims, mutate func(*jwt.Token)) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID
	token.Header["typ"] = tokenType
	if mutate != nil {
		mutate(token)
	}
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func newTestExtension(t *testing.T, contents *atomic.Value) (*telemetryJWTAuthExtension, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		value := contents.Load()
		if value == nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write(value.([]byte))
	}))
	extension := newExtension(Config{
		Issuer:   "https://ctl.example.com",
		Audience: defaultAudience,
		JWKSURL:  server.URL,
	}, zap.NewNop())
	require.NoError(t, extension.Start(context.Background(), nil))
	t.Cleanup(func() {
		require.NoError(t, extension.Shutdown(context.Background()))
		server.Close()
	})
	return extension, server
}

func TestAuthenticateAttachesVerifiedPrincipal(t *testing.T) {
	key := testRSAKey(t)
	var contents atomic.Value
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key}))
	extension, _ := newTestExtension(t, &contents)
	principal := testPrincipal()
	now := time.Now().UTC().Truncate(time.Second)
	raw := signTestToken(t, key, "key-1", testClaims(now, principal), nil)

	ctx, err := extension.Authenticate(context.Background(), map[string][]string{
		"Authorization": {"Bearer " + raw},
	})

	require.NoError(t, err)
	authData, ok := client.FromContext(ctx).Auth.(*AuthData)
	require.True(t, ok)
	require.Equal(t, principal, authData.Principal())
	require.NotContains(t, authData.GetAttributeNames(), "raw")
}

func TestAuthenticateRejectsInvalidTokens(t *testing.T) {
	key := testRSAKey(t)
	var contents atomic.Value
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key}))
	extension, _ := newTestExtension(t, &contents)
	principal := testPrincipal()
	now := time.Now().UTC().Truncate(time.Second)

	tests := map[string]struct {
		mutateClaims func(*telemetryClaims)
		mutateToken  func(*jwt.Token)
	}{
		"wrong audience": {
			mutateClaims: func(claims *telemetryClaims) { claims.Audience = jwt.ClaimStrings{"other"} },
		},
		"extra audience": {
			mutateClaims: func(claims *telemetryClaims) { claims.Audience = append(claims.Audience, "other") },
		},
		"wrong scope": {
			mutateClaims: func(claims *telemetryClaims) { claims.Scope = "other" },
		},
		"missing not before": {
			mutateClaims: func(claims *telemetryClaims) { claims.NotBefore = nil },
		},
		"excessive lifetime": {
			mutateClaims: func(claims *telemetryClaims) {
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(tokenLifetime + time.Minute))
			},
		},
		"wrong client": {
			mutateClaims: func(claims *telemetryClaims) { claims.ClientID = "run" + strings.Repeat("e", 23) },
		},
		"invalid identity": {
			mutateClaims: func(claims *telemetryClaims) { claims.InstallID = "invalid" },
		},
		"wrong subject": {
			mutateClaims: func(claims *telemetryClaims) { claims.Subject = "other" },
		},
		"wrong type": {
			mutateToken: func(token *jwt.Token) { token.Header["typ"] = "JWT" },
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			claims := testClaims(now, principal)
			if test.mutateClaims != nil {
				test.mutateClaims(claims)
			}
			raw := signTestToken(t, key, "key-1", claims, test.mutateToken)

			_, err := extension.Authenticate(context.Background(), map[string][]string{
				"authorization": {"Bearer " + raw},
			})

			require.ErrorIs(t, err, errAuthenticationFailed)
		})
	}
}

func TestAuthenticateRejectsAmbiguousAuthorization(t *testing.T) {
	key := testRSAKey(t)
	var contents atomic.Value
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key}))
	extension, _ := newTestExtension(t, &contents)

	for name, sources := range map[string]map[string][]string{
		"missing": {},
		"multiple values": {
			"authorization": {"Bearer one", "Bearer two"},
		},
		"multiple keys": {
			"authorization": {"Bearer one"},
			"Authorization": {"Bearer two"},
		},
		"wrong scheme": {
			"authorization": {"Basic value"},
		},
		"oversized token": {
			"authorization": {"Bearer " + strings.Repeat("a", maxAccessTokenSize+1) + ".b.c"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := extension.Authenticate(context.Background(), sources)
			require.ErrorIs(t, err, errAuthenticationFailed)
		})
	}
}

func TestStartRejectsJWKSRedirect(t *testing.T) {
	key := testRSAKey(t)
	destination := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key}))
	}))
	t.Cleanup(destination.Close)
	redirect := httptest.NewServer(http.RedirectHandler(destination.URL, http.StatusFound))
	t.Cleanup(redirect.Close)

	extension := newExtension(Config{
		Issuer:   "https://ctl.example.com",
		Audience: defaultAudience,
		JWKSURL:  redirect.URL,
	}, zap.NewNop())

	require.ErrorIs(t, extension.Start(context.Background(), nil), errJWKSUnavailable)
}

func TestAuthenticateRefreshesJWKSForUnknownKey(t *testing.T) {
	key1 := testRSAKey(t)
	key2 := testRSAKey(t)
	var contents atomic.Value
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key1}))
	extension, _ := newTestExtension(t, &contents)
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key1, "key-2": key2}))

	principal := testPrincipal()
	raw := signTestToken(t, key2, "key-2", testClaims(time.Now().UTC().Truncate(time.Second), principal), nil)
	ctx, err := extension.Authenticate(context.Background(), map[string][]string{
		"authorization": {"Bearer " + raw},
	})

	require.NoError(t, err)
	require.Equal(t, principal, client.FromContext(ctx).Auth.(*AuthData).Principal())
}

func TestKeyCacheBoundsKnownKeyStaleness(t *testing.T) {
	key := testRSAKey(t)
	var contents atomic.Value
	contents.Store(testJWKS(t, map[string]*rsa.PrivateKey{"key-1": key}))
	extension, _ := newTestExtension(t, &contents)
	baseTime := time.Now()
	extension.keys.now = func() time.Time { return baseTime.Add(30 * time.Minute) }
	contents.Store([]byte("{"))

	_, err := extension.keys.key(context.Background(), "key-1")
	require.NoError(t, err)

	extension.keys.now = func() time.Time { return baseTime.Add(maxJWKSStaleness + time.Minute) }
	_, err = extension.keys.key(context.Background(), "key-1")
	require.ErrorIs(t, err, errJWKSUnavailable)
}

func TestConfigRequiresHTTPSExceptForLoopback(t *testing.T) {
	for _, test := range []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "HTTPS", url: "https://ctl.example.com"},
		{name: "localhost", url: "http://localhost:8081"},
		{name: "loopback", url: "http://127.0.0.1:8081"},
		{name: "remote HTTP", url: "http://ctl.example.com", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateHTTPSOrLoopbackURL("issuer", test.url)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
