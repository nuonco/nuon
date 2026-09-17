package service

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const (
	telemetryTokenAudience = "urn:nuon:telemetry"
	telemetryTokenScope    = "telemetry:write"
	telemetryTokenLifetime = 10 * time.Minute
	maxTelemetryJWKSSize   = 64 * 1024
)

type TelemetryJSONWebKey struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

type TelemetryJSONWebKeySet struct {
	Keys []TelemetryJSONWebKey `json:"keys"`
}

type telemetryAccessTokenClaims struct {
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
	OrgID     string `json:"nuon_org_id"`
	AppID     string `json:"nuon_app_id"`
	InstallID string `json:"nuon_install_id"`
	RunnerID  string `json:"nuon_runner_id"`
	jwt.RegisteredClaims
}

type telemetryTokenIssuer struct {
	issuer     string
	keyID      string
	privateKey *rsa.PrivateKey
	publicKeys TelemetryJSONWebKeySet
	now        func() time.Time
}

type telemetryJSONWebKeySet struct {
	Keys []telemetryJSONWebKey `json:"keys"`
}

type telemetryJSONWebKey struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
	D         string `json:"d"`
	P         string `json:"p"`
	Q         string `json:"q"`
	DP        string `json:"dp"`
	DQ        string `json:"dq"`
	QI        string `json:"qi"`
}

func newTelemetryTokenIssuer(cfg *internal.Config) (*telemetryTokenIssuer, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, nil
	}
	if len(cfg.TelemetryJWKS) > maxTelemetryJWKSSize {
		return nil, fmt.Errorf("telemetry JWKS exceeds maximum size")
	}

	issuer := strings.TrimRight(cfg.TelemetryJWTIssuer, "/")
	if issuer == "" {
		issuer = strings.TrimRight(cfg.PublicAPIURL, "/")
	}
	if err := validateTelemetryIssuer(issuer); err != nil {
		return nil, err
	}

	privateKey, keyID, publicKeys, err := parseTelemetryJWKS(cfg.TelemetryJWKS)
	if err != nil {
		return nil, err
	}

	return &telemetryTokenIssuer{
		issuer:     issuer,
		keyID:      keyID,
		privateKey: privateKey,
		publicKeys: publicKeys,
		now:        time.Now,
	}, nil
}

func validateTelemetryIssuer(issuer string) error {
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("telemetry JWT issuer must be an absolute HTTP or HTTPS URL without userinfo, query, or fragment")
	}
	return nil
}

func parseTelemetryJWKS(value string) (*rsa.PrivateKey, string, TelemetryJSONWebKeySet, error) {
	var input telemetryJSONWebKeySet
	if err := json.Unmarshal([]byte(value), &input); err != nil {
		return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("decode telemetry JWKS: %w", err)
	}
	if len(input.Keys) == 0 {
		return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWKS contains no keys")
	}

	publicKeys := TelemetryJSONWebKeySet{Keys: make([]TelemetryJSONWebKey, 0, len(input.Keys))}
	seenKeyIDs := make(map[string]struct{}, len(input.Keys))
	var signingKey *rsa.PrivateKey
	var signingKeyID string
	for _, key := range input.Keys {
		if key.KeyID == "" {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWK key ID is required")
		}
		if _, exists := seenKeyIDs[key.KeyID]; exists {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWKS contains duplicate key IDs")
		}
		seenKeyIDs[key.KeyID] = struct{}{}

		if key.KeyType != "RSA" || (key.Use != "" && key.Use != "sig") || (key.Algorithm != "" && key.Algorithm != jwt.SigningMethodRS256.Alg()) {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWKS supports only RSA signing keys using RS256")
		}

		publicKey, err := parseTelemetryRSAPublicKey(key)
		if err != nil {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("parse telemetry JWK %q: %w", key.KeyID, err)
		}
		publicKeys.Keys = append(publicKeys.Keys, TelemetryJSONWebKey{
			KeyType:   "RSA",
			KeyID:     key.KeyID,
			Use:       "sig",
			Algorithm: jwt.SigningMethodRS256.Alg(),
			Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			Exponent:  encodeTelemetryJWKInteger(int64(publicKey.E)),
		})

		if key.D == "" && key.P == "" && key.Q == "" && key.DP == "" && key.DQ == "" && key.QI == "" {
			continue
		}
		if signingKey != nil {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWKS must contain exactly one private signing key")
		}

		signingKey, err = parseTelemetryRSAPrivateKey(key, publicKey)
		if err != nil {
			return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("parse telemetry private JWK %q: %w", key.KeyID, err)
		}
		signingKeyID = key.KeyID
	}

	if signingKey == nil {
		return nil, "", TelemetryJSONWebKeySet{}, fmt.Errorf("telemetry JWKS does not contain a private signing key")
	}
	return signingKey, signingKeyID, publicKeys, nil
}

func parseTelemetryRSAPublicKey(key telemetryJSONWebKey) (*rsa.PublicKey, error) {
	modulus, err := decodeTelemetryJWKInteger(key.Modulus)
	if err != nil || modulus.Sign() <= 0 {
		return nil, fmt.Errorf("invalid RSA modulus")
	}
	if modulus.BitLen() < 2048 {
		return nil, fmt.Errorf("RSA modulus must be at least 2048 bits")
	}
	exponent, err := decodeTelemetryJWKInteger(key.Exponent)
	if err != nil || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > int64(^uint(0)>>1) || exponent.Int64()%2 == 0 {
		return nil, fmt.Errorf("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

func parseTelemetryRSAPrivateKey(key telemetryJSONWebKey, publicKey *rsa.PublicKey) (*rsa.PrivateKey, error) {
	if key.D == "" || key.P == "" || key.Q == "" {
		return nil, fmt.Errorf("RSA private key requires d, p, and q")
	}
	d, err := decodeTelemetryJWKInteger(key.D)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA private exponent")
	}
	p, err := decodeTelemetryJWKInteger(key.P)
	if err != nil {
		return nil, fmt.Errorf("invalid first RSA prime")
	}
	q, err := decodeTelemetryJWKInteger(key.Q)
	if err != nil {
		return nil, fmt.Errorf("invalid second RSA prime")
	}

	privateKey := &rsa.PrivateKey{
		PublicKey: *publicKey,
		D:         d,
		Primes:    []*big.Int{p, q},
	}
	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("validate RSA private key: %w", err)
	}
	privateKey.Precompute()
	return privateKey, nil
}

func decodeTelemetryJWKInteger(value string) (*big.Int, error) {
	if value == "" {
		return nil, errors.New("JWK integer is empty")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 {
		return nil, errors.New("JWK integer is not valid base64url")
	}
	return new(big.Int).SetBytes(decoded), nil
}

func encodeTelemetryJWKInteger(value int64) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(value).Bytes())
}

func (i *telemetryTokenIssuer) issue(principal telemetryRunnerPrincipal) (string, error) {
	if principal.OrgID == "" || principal.AppID == "" || principal.InstallID == "" || principal.RunnerID == "" {
		return "", fmt.Errorf("telemetry runner principal is incomplete")
	}

	now := i.now().UTC()
	claims := telemetryAccessTokenClaims{
		ClientID:  principal.RunnerID,
		Scope:     telemetryTokenScope,
		OrgID:     principal.OrgID,
		AppID:     principal.AppID,
		InstallID: principal.InstallID,
		RunnerID:  principal.RunnerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   fmt.Sprintf("org:%s:install:%s:runner:%s", principal.OrgID, principal.InstallID, principal.RunnerID),
			Audience:  jwt.ClaimStrings{telemetryTokenAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(telemetryTokenLifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.keyID
	token.Header["typ"] = "at+jwt"

	signed, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign telemetry access token: %w", err)
	}
	return signed, nil
}

func (i *telemetryTokenIssuer) publicJWKS() TelemetryJSONWebKeySet {
	keys := make([]TelemetryJSONWebKey, len(i.publicKeys.Keys))
	copy(keys, i.publicKeys.Keys)
	return TelemetryJSONWebKeySet{Keys: keys}
}
