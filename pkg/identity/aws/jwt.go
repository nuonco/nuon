package aws

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// unverifiedJWT represents a JWT token that has been parsed but not verified
type unverifiedJWT struct {
	Header  map[string]interface{}
	Payload map[string]interface{}
	Issuer  string
}

// parseUnverifiedJWT parses a JWT token without verifying its signature
// This is used to extract metadata (like region) before validation
func parseUnverifiedJWT(token string) (*unverifiedJWT, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT header: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse JWT header: %w", err)
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JWT payload: %w", err)
	}

	// Extract issuer
	issuer, _ := payload["iss"].(string)

	return &unverifiedJWT{
		Header:  header,
		Payload: payload,
		Issuer:  issuer,
	}, nil
}
