package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/browser"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var openBrowser = browser.OpenURL

const (
	nuonAuthDeviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"
	nuonAuthDefaultPollInterval = 5 * time.Second
	nuonAuthSlowDownIncrement   = 5 * time.Second

	// loopbackMinCtlAPIVersion is the first control-plane release that serves
	// the CLI authorization-code grant. Older releases only have device login.
	loopbackMinCtlAPIVersion = "0.19.1289"
)

// NuonAuthDeviceAuthorizationResponse is the device authorization response (RFC 8628 §3.2).
type NuonAuthDeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// NuonAuthTokenResponse represents the response from the device token endpoint
type NuonAuthTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Email            string `json:"email,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// buildAuthURL constructs the auth service URL based on the root domain
func buildAuthURL(rootDomain string) string {
	if rootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://auth.%s", rootDomain)
}

func (a *Service) nuonAuthBaseURL(rootDomain string) string {
	if a.authBaseURLOverride != "" {
		return a.authBaseURLOverride
	}
	return buildAuthURL(rootDomain)
}

func (a *Service) loopbackLoginEnabled() bool {
	return a.cfg != nil && a.cfg.Interactive
}

// loginWithNuonAuth signs the CLI in. An interactive terminal uses the
// authorization-code grant with a 127.0.0.1 callback when this control plane
// has it. Anything else uses the device authorization grant.
func (a *Service) loginWithNuonAuth(ctx context.Context, cfg *models.ServiceCLIConfig) (*LoginResult, error) {
	authBaseURL := a.nuonAuthBaseURL(cfg.RootDomain)
	if a.loopbackLoginEnabled() && a.ctlAPISupportsLoopback(ctx) {
		result, err := a.loginWithLoopback(ctx, authBaseURL)
		if !errors.Is(err, errLoopbackUnavailable) {
			return result, err
		}
	}
	return a.loginWithDeviceCode(ctx, authBaseURL)
}

// ctlAPISupportsLoopback reports whether this control plane serves the CLI
// loopback grant. A release older than 0.19.1289 does not. A value that is
// not a release number, such as the git hash local and stage builds report,
// is a current build and does. A failed lookup uses device login, which both
// generations serve.
func (a *Service) ctlAPISupportsLoopback(ctx context.Context) bool {
	if a.api == nil {
		return true
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	remote, err := a.api.GetVersion(ctx)
	if err != nil || remote == nil {
		return false
	}
	return loopbackSupportedByCtlAPI(remote.Version)
}

func loopbackSupportedByCtlAPI(version string) bool {
	parsed, err := semver.NewVersion(strings.TrimSpace(version))
	if err != nil {
		return true
	}
	minimum, err := semver.NewVersion(loopbackMinCtlAPIVersion)
	if err != nil {
		return true
	}
	return !parsed.LessThan(minimum)
}

// loginWithDeviceCode performs the device authorization grant (RFC 8628) against the Nuon auth service.
func (a *Service) loginWithDeviceCode(ctx context.Context, authBaseURL string) (*LoginResult, error) {
	authz, err := startNuonDeviceAuthorization(ctx, authBaseURL)
	if err != nil {
		return nil, err
	}

	openURL := authz.VerificationURIComplete
	if openURL == "" {
		openURL = authz.VerificationURI
	}

	fmt.Println("\nLogging in to Nuon")
	fmt.Printf("Opening your browser. Check that it shows this code:\n\n    %s\n\n", authz.UserCode)
	fmt.Printf("If the browser does not open, visit %s and enter the code.\n", authz.VerificationURI)
	fmt.Println("Never approve a login code that someone else sent you.")

	_ = openBrowser(openURL)

	interval := time.Duration(authz.Interval) * time.Second
	if interval <= 0 {
		interval = nuonAuthDefaultPollInterval
	}
	deadline := time.Now().Add(time.Duration(authz.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		resp, err := pollNuonDeviceToken(ctx, authBaseURL, authz.DeviceCode)
		if err != nil {
			return nil, err
		}

		if resp.AccessToken != "" {
			fmt.Println()
			displayName := resp.Email
			if displayName == "" {
				displayName = "user"
			}
			return &LoginResult{
				AccessToken: resp.AccessToken,
				DisplayName: displayName,
			}, nil
		}

		switch resp.Error {
		case "authorization_pending":
			fmt.Print(".")
			continue
		case "slow_down":
			interval += nuonAuthSlowDownIncrement
			continue
		case "expired_token":
			fmt.Println()
			return nil, fmt.Errorf("login code expired - please try again")
		case "access_denied":
			fmt.Println()
			return nil, fmt.Errorf("access denied")
		default:
			fmt.Println()
			return nil, fmt.Errorf("authentication error: %s - %s", resp.Error, resp.ErrorDescription)
		}
	}

	fmt.Println()
	return nil, fmt.Errorf("login code expired - please try again")
}

// startNuonDeviceAuthorization asks the auth service to issue a device code and user code.
func startNuonDeviceAuthorization(ctx context.Context, authBaseURL string) (*NuonAuthDeviceAuthorizationResponse, error) {
	body, status, err := postNuonAuthForm(ctx, authBaseURL+"/device/authorize", url.Values{})
	if err != nil {
		return nil, fmt.Errorf("failed to start login: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to start login: auth service returned %d", status)
	}

	var authz NuonAuthDeviceAuthorizationResponse
	if err := json.Unmarshal(body, &authz); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}
	if authz.DeviceCode == "" || authz.UserCode == "" || authz.VerificationURI == "" {
		return nil, fmt.Errorf("auth service returned an incomplete login response")
	}
	return &authz, nil
}

// pollNuonDeviceToken makes a single request to the token endpoint. Per RFC 8628 pending and error
// states come back as 400s with an error code, so the body is parsed regardless of status.
func pollNuonDeviceToken(ctx context.Context, authBaseURL, deviceCode string) (*NuonAuthTokenResponse, error) {
	body, status, err := postNuonAuthForm(ctx, authBaseURL+"/device/token", url.Values{
		"grant_type":  {nuonAuthDeviceCodeGrantType},
		"device_code": {deviceCode},
	})
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

	var tokenResp NuonAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response (status %d): %w", status, err)
	}
	return &tokenResp, nil
}

func postNuonAuthForm(ctx context.Context, endpoint string, form url.Values) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}
	return body, resp.StatusCode, nil
}
