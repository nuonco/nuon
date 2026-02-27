package service

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
// This is used to extract metadata (like region or tenant ID) before validation
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

// extractRegionFromAWSIssuer extracts the AWS region from an STS issuer URL
// Example: https://sts.us-west-2.amazonaws.com -> us-west-2
func extractRegionFromAWSIssuer(issuer string) (string, error) {
	// Remove https:// prefix
	host := strings.TrimPrefix(issuer, "https://")
	host = strings.TrimPrefix(host, "http://")

	// Parse format: sts.{region}.amazonaws.com or sts.amazonaws.com (global)
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid AWS STS issuer format: %s", issuer)
	}

	// Global STS endpoint (no region)
	if parts[0] == "sts" && parts[1] == "amazonaws" {
		return "us-east-1", nil // Default to us-east-1 for global endpoint
	}

	// Regional endpoint
	if parts[0] == "sts" && len(parts) >= 3 {
		return parts[1], nil
	}

	return "", fmt.Errorf("unable to extract region from AWS issuer: %s", issuer)
}

// extractTenantFromAzureIssuer extracts the tenant ID from an Azure AD issuer URL
// Example: https://login.microsoftonline.com/{tenant}/v2.0 -> {tenant}
func extractTenantFromAzureIssuer(issuer string) (string, error) {
	// Remove https:// prefix
	path := strings.TrimPrefix(issuer, "https://")
	path = strings.TrimPrefix(path, "http://")

	// Remove domain
	path = strings.TrimPrefix(path, "login.microsoftonline.com/")
	path = strings.TrimPrefix(path, "sts.windows.net/")

	// Remove /v2.0 or /v1.0 suffix
	path = strings.TrimSuffix(path, "/v2.0")
	path = strings.TrimSuffix(path, "/v1.0")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return "", fmt.Errorf("unable to extract tenant ID from Azure issuer: %s", issuer)
	}

	return path, nil
}
