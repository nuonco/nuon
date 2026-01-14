package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	nuonAuthDeviceCodeExpiry = 5 * time.Minute
	nuonAuthPollInterval     = 5 * time.Second
)

// NuonAuthTokenResponse represents the response from the device token endpoint
type NuonAuthTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Email            string `json:"email,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// generateDeviceCode creates a device code in format XXXX-XXXX where X is [A-Z0-9]
func generateDeviceCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate device code: %w", err)
	}

	// Convert random bytes to characters from our allowed set
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}

	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", string(b[:4]), string(b[4:])), nil
}

// buildAuthURL constructs the auth service URL based on the root domain
func buildAuthURL(rootDomain string) string {
	if rootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://auth.%s", rootDomain)
}

// loginWithNuonAuth performs the Nuon Auth device code flow
func (a *Service) loginWithNuonAuth(ctx context.Context, cfg *models.ServiceCLIConfig) (*LoginResult, error) {
	// Generate device code locally
	deviceCode, err := generateDeviceCode()
	if err != nil {
		return nil, err
	}

	// Build URLs based on root domain
	authBaseURL := buildAuthURL(cfg.RootDomain)
	verificationURL := fmt.Sprintf("%s/device/code?code=%s", authBaseURL, deviceCode)
	tokenURL := fmt.Sprintf("%s/device/token?code=%s", authBaseURL, deviceCode)

	// Display instructions
	fmt.Println("\nLogging in to Nuon")
	fmt.Printf("Opening your browser to authorize the CLI.\n")
	fmt.Printf("If the browser does not open, visit this URL:\n\n%s\n\n", verificationURL)

	// Open browser
	browser.OpenURL(verificationURL)

	// Poll for token with timeout
	deadline := time.Now().Add(nuonAuthDeviceCodeExpiry)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(nuonAuthPollInterval):
		}

		resp, err := a.pollNuonDeviceToken(tokenURL)
		if err != nil {
			return nil, err
		}

		// Check for success
		if resp.AccessToken != "" {
			fmt.Println() // newline after dots
			displayName := resp.Email
			if displayName == "" {
				displayName = "user"
			}
			return &LoginResult{
				AccessToken: resp.AccessToken,
				DisplayName: displayName,
			}, nil
		}

		// Check error type
		switch resp.Error {
		case "authorization_pending":
			fmt.Print(".")
			continue
		case "expired_token":
			fmt.Println()
			return nil, fmt.Errorf("device code expired - please try again")
		case "access_denied":
			fmt.Println()
			return nil, fmt.Errorf("access denied")
		case "invalid_code":
			fmt.Println()
			return nil, fmt.Errorf("invalid device code: %s", resp.ErrorDescription)
		case "slow_down":
			time.Sleep(nuonAuthPollInterval)
			continue
		default:
			if resp.Error != "" {
				fmt.Println()
				return nil, fmt.Errorf("authentication error: %s - %s", resp.Error, resp.ErrorDescription)
			}
		}
	}

	fmt.Println()
	return nil, fmt.Errorf("device code expired - please try again")
}

// pollNuonDeviceToken makes a single request to the token endpoint
func (a *Service) pollNuonDeviceToken(tokenURL string) (*NuonAuthTokenResponse, error) {
	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Network errors during polling are not fatal - could be temporary
		if strings.Contains(err.Error(), "connection refused") {
			return &NuonAuthTokenResponse{
				Error:            "authorization_pending",
				ErrorDescription: "waiting for auth service",
			}, nil
		}
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp NuonAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenResp, nil
}
