package auth

import "time"

// OIDCAuthRequest represents a request to authenticate a runner via OIDC
type OIDCAuthRequest struct {
	Provider string `json:"provider" validate:"required,oneof=aws gcp azure"`
	Token    string `json:"token" validate:"required"`
	RunnerID string `json:"runner_id" validate:"required"`
}

// OIDCAuthResponse represents the response from OIDC authentication
type OIDCAuthResponse struct {
	Authenticated bool      `json:"authenticated"`
	Provider      string    `json:"provider"`
	RunnerID      string    `json:"runner_id"`
	Token         string    `json:"token"`
	ExpiresAt     time.Time `json:"expires_at"`
}
