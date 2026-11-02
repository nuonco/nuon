package acr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/jwtcompact"
)

type JWTClaims struct {
	Audience   string `json:"aud"`
	Issuer     string `json:"iss"`
	TenantID   string `json:"tid"`
	Subject    string `json:"sub"`
	Expiration int64  `json:"exp"`
}

// parseJWT parses a JWT token from Azure to get metadata about it.
func parseJWT(tokenString string) (*JWTClaims, error) {
	if err := jwtcompact.Require(tokenString); err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(tokenString), ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}
