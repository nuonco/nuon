package oidcissuer

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func testRSAPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func testPrivateJWKS(t *testing.T, key *rsa.PrivateKey, keyID string) string {
	t.Helper()

	contents, err := json.Marshal(privateJWKSet{Keys: []privateJWK{{
		KeyType:   "RSA",
		KeyID:     keyID,
		Use:       "sig",
		Algorithm: jwt.SigningMethodRS256.Alg(),
		Modulus:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		Exponent:  encodeJWKInteger(int64(key.E)),
		D:         base64.RawURLEncoding.EncodeToString(key.D.Bytes()),
		P:         base64.RawURLEncoding.EncodeToString(key.Primes[0].Bytes()),
		Q:         base64.RawURLEncoding.EncodeToString(key.Primes[1].Bytes()),
	}}})
	require.NoError(t, err)
	return string(contents)
}

func newTestIssuer(t *testing.T, key *rsa.PrivateKey, keyID string, now time.Time) *Issuer {
	t.Helper()

	issuer, err := New("https://ctl.example.com", key, keyID)
	require.NoError(t, err)
	issuer.now = func() time.Time { return now }
	return issuer
}

func publishedJWKSKey(t *testing.T, published JWKS) *rsa.PublicKey {
	t.Helper()

	require.Len(t, published.Keys, 1)
	key := published.Keys[0]
	require.Equal(t, "RSA", key.KeyType)
	require.Equal(t, "sig", key.Use)
	require.Equal(t, "RS256", key.Algorithm)

	modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	require.NoError(t, err)
	exponent, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	require.NoError(t, err)
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(new(big.Int).SetBytes(exponent).Int64()),
	}
}

func TestIssuerMint(t *testing.T) {
	const (
		keyID      = "test-key-1"
		subjectFmt = "org:org-test:connection:conn-test"
	)
	now := time.Date(2026, time.September, 24, 12, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		subject  string
		audience string
		ttl      time.Duration
		wantErr  string
	}{
		"aws federation": {
			subject:  subjectFmt,
			audience: "https://sts.amazonaws.com",
		},
		"azure federation": {
			subject:  subjectFmt,
			audience: "api://AzureADTokenExchange",
		},
		"gcp federation": {
			subject:  subjectFmt,
			audience: "https://iam.googleapis.com/projects/example-project/locations/global/workloadIdentityPools/example-pool/providers/example-provider",
		},
		"missing subject": {
			audience: "https://sts.amazonaws.com",
			wantErr:  "subject is required",
		},
		"missing audience": {
			subject: subjectFmt,
			wantErr: "audience is required",
		},
		"zero ttl": {
			subject:  subjectFmt,
			audience: "https://sts.amazonaws.com",
			ttl:      0,
			wantErr:  "TTL must be positive",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			key := testRSAPrivateKey(t)
			issuer := newTestIssuer(t, key, keyID, now)

			tokenTTL := test.ttl
			if tokenTTL == 0 && test.wantErr == "" {
				tokenTTL = time.Minute
			}
			token, err := issuer.Mint(t.Context(), test.subject, test.audience, tokenTTL)
			if test.wantErr != "" {
				require.Empty(t, token)
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)

			// Verify the signature against the published JWKS, not the
			// private key, so a wrong or missing published key fails here.
			_, keyID, published, err := ParseJWKS(testPrivateJWKS(t, key, keyID))
			require.NoError(t, err)

			claims := &jwt.RegisteredClaims{}
			parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
				return publishedJWKSKey(t, published), nil
			},
				jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
				jwt.WithIssuer(issuer.Issuer()),
				jwt.WithAudience(test.audience),
				jwt.WithTimeFunc(func() time.Time { return now }),
			)
			require.NoError(t, err)
			require.True(t, parsed.Valid)
			require.Equal(t, keyID, parsed.Header["kid"])
			require.Equal(t, "JWT", parsed.Header["typ"])

			require.Equal(t, "https://ctl.example.com", claims.Issuer)
			require.Equal(t, test.subject, claims.Subject)
			require.Equal(t, jwt.ClaimStrings{test.audience}, claims.Audience)
			require.True(t, now.Equal(claims.IssuedAt.Time))
			require.True(t, now.Equal(claims.NotBefore.Time))
			require.True(t, now.Add(tokenTTL).Equal(claims.ExpiresAt.Time))
			require.NotEmpty(t, claims.ID)
		})
	}
}

func TestIssuerDiscoveryDocumentMatchesIssuer(t *testing.T) {
	key := testRSAPrivateKey(t)
	issuer, err := New("https://ctl.example.com/", key, "test-key-1")
	require.NoError(t, err)

	document := issuer.DiscoveryDocument()
	require.Equal(t, "https://ctl.example.com", document.Issuer)
	require.Equal(t, issuer.Issuer(), document.Issuer)
	require.Equal(t, "https://ctl.example.com/.well-known/jwks.json", document.JWKSURI)
	require.Equal(t, issuer.KeyID(), "test-key-1")
	require.Equal(t, []string{"id_token"}, document.ResponseTypesSupported)
	require.Equal(t, []string{"public"}, document.SubjectTypesSupported)
	require.Equal(t, []string{"RS256"}, document.IDTokenSigningAlgValuesSupported)
	require.Equal(t, []string{"iss", "sub", "aud", "iat", "exp", "nbf", "jti"}, document.ClaimsSupported)
}
