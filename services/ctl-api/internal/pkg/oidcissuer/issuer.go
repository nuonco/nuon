package oidcissuer

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// JWKSPath is the well-known path where the control plane publishes its
	// public keys; it must stay in sync with the route registered on the
	// public listener.
	JWKSPath = "/.well-known/jwks.json"
)

// Issuer mints short-lived RS256 federation and access tokens signed with
// the control plane's signing key.
type Issuer struct {
	issuer     string
	keyID      string
	privateKey *rsa.PrivateKey
	now        func() time.Time
}

func New(issuer string, privateKey *rsa.PrivateKey, keyID string) (*Issuer, error) {
	issuer = strings.TrimRight(issuer, "/")
	if err := ValidateIssuer(issuer); err != nil {
		return nil, err
	}
	if privateKey == nil {
		return nil, fmt.Errorf("signing key is required")
	}
	if keyID == "" {
		return nil, fmt.Errorf("signing key ID is required")
	}

	return &Issuer{
		issuer:     issuer,
		keyID:      keyID,
		privateKey: privateKey,
		now:        time.Now,
	}, nil
}

func NewFromJWKS(issuerURL, keySet string) (*Issuer, JWKS, error) {
	privateKey, keyID, publicKeys, err := ParseJWKS(keySet)
	if err != nil {
		return nil, JWKS{}, err
	}
	issuer, err := New(issuerURL, privateKey, keyID)
	if err != nil {
		return nil, JWKS{}, err
	}
	return issuer, publicKeys, nil
}

// Mint returns a signed token with iss, sub, aud, iat, exp, nbf, and jti
// claims, signed with the configured key. The issuer matches the discovery
// document exactly, as OIDC federation requires.
func (i *Issuer) Mint(ctx context.Context, subject, audience string, ttl time.Duration) (string, error) {
	return i.mint(subject, audience, ttl, "JWT", nil)
}

func (i *Issuer) MintAccessToken(subject, audience string, ttl time.Duration, customClaims jwt.MapClaims) (string, error) {
	return i.mint(subject, audience, ttl, "at+jwt", customClaims)
}

func (i *Issuer) mint(subject, audience string, ttl time.Duration, tokenType string, customClaims jwt.MapClaims) (string, error) {
	if subject == "" {
		return "", fmt.Errorf("OIDC token subject is required")
	}
	if audience == "" {
		return "", fmt.Errorf("OIDC token audience is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("OIDC token TTL must be positive")
	}

	claims := make(jwt.MapClaims, len(customClaims)+7)
	for name, value := range customClaims {
		switch name {
		case "iss", "sub", "aud", "iat", "exp", "nbf", "jti":
			return "", fmt.Errorf("custom claims cannot override %q", name)
		}
		claims[name] = value
	}

	now := i.now().UTC()
	claims["iss"] = i.issuer
	claims["sub"] = subject
	claims["aud"] = jwt.ClaimStrings{audience}
	claims["exp"] = jwt.NewNumericDate(now.Add(ttl))
	claims["nbf"] = jwt.NewNumericDate(now)
	claims["iat"] = jwt.NewNumericDate(now)
	claims["jti"] = uuid.NewString()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.keyID
	token.Header["typ"] = tokenType

	signed, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign OIDC token: %w", err)
	}
	return signed, nil
}

func (i *Issuer) Issuer() string {
	return i.issuer
}

func (i *Issuer) KeyID() string {
	return i.keyID
}

type DiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

func (i *Issuer) DiscoveryDocument() DiscoveryDocument {
	return DiscoveryDocument{
		Issuer:                           i.issuer,
		JWKSURI:                          i.issuer + JWKSPath,
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{jwt.SigningMethodRS256.Alg()},
		ClaimsSupported:                  []string{"iss", "sub", "aud", "iat", "exp", "nbf", "jti"},
	}
}
