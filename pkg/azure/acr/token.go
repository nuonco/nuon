package acr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/redact"
)

const (
	DefaultACRUsername string = "00000000-0000-0000-0000-000000000000"
	// RegistryScope is the Entra scope exchanged for an ACR refresh token.
	// That access token is posted to the registry login server, so it is limited to Container Registry.
	RegistryScope = "https://containerregistry.azure.net/.default"
	armTokenScope = "https://management.azure.com/.default"
)

// exchangeClient never follows redirects: the request body carries an Entra
// token, and a redirect would replay it to whatever host the registry names.
var exchangeClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var ConnectionExchangeClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
			Control: func(_, address string, _ syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				ip := net.ParseIP(host)
				if ip == nil {
					return fmt.Errorf("acr token exchange refused non-IP address %q", host)
				}
				if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
					return fmt.Errorf("acr token exchange refused address %q", address)
				}
				return nil
			},
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// GetRepositoryToken exchanges an Azure credential for a refresh token that can be used to authenticate with
// the registry. It has a timeout of 60 minutes.
//
// cfg selects the identity: an app registration when it carries one (the only
// way to reach a registry in another tenant), otherwise the ambient identity.
// A nil cfg means ambient.
//
// NOTE: we do this, instead of using the ACR repository client to simplify our dependencies, however, at some point we
// plan on moving this into a package, like we have with `pkg/aws`.
func GetRepositoryToken(ctx context.Context, cfg *credentials.Config, acrService string, logger *zap.Logger) (string, error) {
	// get a credential
	credential, err := credentials.Fetch(ctx, cfg, logger)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	tenantID := ""
	if cfg != nil {
		tenantID = cfg.TenantID
	}
	return exchangeRepositoryToken(ctx, credential, tenantID, acrService, armTokenScope, exchangeClient)
}

func GetRepositoryTokenWithCredential(ctx context.Context, credential azcore.TokenCredential, tenantID, acrService string) (string, error) {
	return exchangeRepositoryToken(ctx, credential, tenantID, acrService, RegistryScope, ConnectionExchangeClient)
}

func exchangeRepositoryToken(ctx context.Context, credential azcore.TokenCredential, tenantID, acrService, scope string, client *http.Client) (string, error) {
	if err := ValidateLoginServer(acrService); err != nil {
		return "", err
	}
	aadToken, err := credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{scope}},
	)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}

	claims, err := parseJWT(aadToken.Token)
	if err != nil {
		return "", fmt.Errorf("unable to parse entra id token for claims: %w", err)
	}

	if tenantID == "" {
		tenantID = claims.TenantID
	}

	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrService},
		"tenant":       {tenantID},
		"access_token": {aadToken.Token},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://%s/oauth2/exchange", acrService), strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("unable to build token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	jsonResponse, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get credential: %w", err)
	}
	defer jsonResponse.Body.Close()

	if jsonResponse.StatusCode != http.StatusOK {
		message := fmt.Sprintf("acr token exchange with %s returned status %d", acrService, jsonResponse.StatusCode)
		if diagnostic := exchangeFailureDiagnostic(jsonResponse.Body, aadToken.Token); diagnostic != "" {
			message += ": " + diagnostic
		}
		return "", errors.New(message)
	}

	var response map[string]interface{}
	decoder := json.NewDecoder(jsonResponse.Body)
	if err := decoder.Decode(&response); err != nil {
		return "", fmt.Errorf("unable to parse token response: %w", err)
	}
	rawToken := response["refresh_token"]
	token, ok := rawToken.(string)
	if !ok {
		return "", fmt.Errorf("unable to parse refresh token as string")
	}

	// return the refresh token
	return token, nil
}

// AccessToken exchanges a refresh token for an access token limited to scope,
// for example repository:org/app:pull,push.
func AccessToken(ctx context.Context, acrService, scope, refreshToken string) (string, error) {
	return getAccessToken(ctx, acrService, scope, refreshToken)
}

const (
	maxExchangeFailureBody = 16 * 1024
	maxExchangeDiagnostic  = 2048
)

func exchangeFailureDiagnostic(body io.Reader, credential string) string {
	data, err := io.ReadAll(io.LimitReader(body, maxExchangeFailureBody+1))
	if err != nil || len(data) > maxExchangeFailureBody {
		return ""
	}
	var failure struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Errors           []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if json.Unmarshal(data, &failure) != nil {
		return ""
	}

	var reasons []string
	addReason := func(code, detail string) {
		if code != "" {
			reasons = append(reasons, strings.TrimSuffix(code+": "+redact.DiagnosticSecrets(detail), ": "))
		}
	}
	addReason(failure.Error, failure.ErrorDescription)
	for _, entry := range failure.Errors {
		addReason(entry.Code, entry.Message)
	}
	diagnostic := strings.Join(reasons, "; ")
	for _, secret := range []string{credential, url.QueryEscape(credential), url.PathEscape(credential)} {
		if secret != "" {
			diagnostic = strings.ReplaceAll(diagnostic, secret, redact.Placeholder)
		}
	}
	diagnostic = strings.Join(strings.Fields(diagnostic), " ")
	if len(diagnostic) > maxExchangeDiagnostic {
		diagnostic = strings.ToValidUTF8(diagnostic[:maxExchangeDiagnostic], "") + "..."
	}
	return diagnostic
}
