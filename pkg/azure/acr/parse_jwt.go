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

func parseJWT(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
