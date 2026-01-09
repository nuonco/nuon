package service

import (
	"fmt"
	"time"

	"gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

const (
	// JWTDefaultExpiry is the default token lifetime.
	JWTDefaultExpiry = 24 * time.Hour

	// JWTIssuer is the issuer claim value.
	JWTIssuer = "nuon-auth"
)

// JWTClaims represents the parsed claims from a Nuon auth JWT.
// This is the public-facing struct returned by parseJWT.
type JWTClaims struct {
	Subject      string
	Email        string
	Username     string
	ProviderType string
}

// nuonAuthClaims represents the custom claims encoded in the JWT.
// This is internal to the JWT encoding/decoding.
type nuonAuthClaims struct {
	Email        string `json:"email,omitempty"`
	Username     string `json:"username,omitempty"`
	ProviderType string `json:"provider_type,omitempty"`
}

// createJWT creates a signed JWT containing user information.
func (s *service) createJWT(userInfo *providers.UserInfo, providerType string) (string, error) {
	// Create the signer using HS256 (HMAC-SHA256)
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: []byte(s.cfg.NuonAuthJWTSecret)},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	// Build the claims
	now := time.Now()
	claims := jwt.Claims{
		Issuer:    JWTIssuer,
		Subject:   userInfo.Subject,
		IssuedAt:  jwt.NewNumericDate(now),
		Expiry:    jwt.NewNumericDate(now.Add(JWTDefaultExpiry)),
		NotBefore: jwt.NewNumericDate(now),
	}

	customClaims := nuonAuthClaims{
		Email:        userInfo.Email,
		Username:     userInfo.Username,
		ProviderType: providerType,
	}

	// Sign the token with both standard and custom claims
	token, err := jwt.Signed(sig).Claims(claims).Claims(customClaims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return token, nil
}

// parseJWT parses and validates a JWT token.
func (s *service) parseJWT(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Verify signature and extract claims
	var claims jwt.Claims
	var customClaims nuonAuthClaims

	if err := token.Claims([]byte(s.cfg.NuonAuthJWTSecret), &claims, &customClaims); err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Validate standard claims (expiry, not before, etc.)
	expected := jwt.Expected{
		Issuer: JWTIssuer,
		Time:   time.Now(),
	}
	if err := claims.Validate(expected); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return &JWTClaims{
		Subject:      claims.Subject,
		Email:        customClaims.Email,
		Username:     customClaims.Username,
		ProviderType: customClaims.ProviderType,
	}, nil
}
