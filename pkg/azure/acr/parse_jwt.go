package acr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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
	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var claims JWTClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
