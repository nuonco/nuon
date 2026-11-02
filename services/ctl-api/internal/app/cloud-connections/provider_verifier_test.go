package cloudconnections

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type verificationTransportFunc func(*http.Request) (*http.Response, error)

func (f verificationTransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGARVerification(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		stsError      string
		impersonation bool
		allowForeign  bool
		foreignError  string
		retry         bool
		failures      int
		wantSleeps    int
		wantStatus    app.CloudConnectionStatus
		wantMessage   string
		wantError     bool
	}{
		"verified":                  {wantStatus: app.CloudConnectionStatusVerified},
		"federation denied":         {stsError: "invalid_grant", wantStatus: app.CloudConnectionStatusError, wantMessage: "invalid_grant: The given credential is rejected by the attribute condition"},
		"impersonation denied":      {impersonation: true, wantStatus: app.CloudConnectionStatusError, wantMessage: "PERMISSION_DENIED: Permission iam.serviceAccounts.getAccessToken denied for acme"},
		"retry exhausted":           {stsError: "invalid_grant", retry: true, wantSleeps: 12, wantStatus: app.CloudConnectionStatusError, wantMessage: "invalid_grant: The given credential is rejected by the attribute condition"},
		"retry succeeds":            {stsError: "invalid_grant", retry: true, failures: 1, wantSleeps: 1, wantStatus: app.CloudConnectionStatusVerified},
		"foreign accepted":          {allowForeign: true, wantStatus: app.CloudConnectionStatusError, wantMessage: "accepts a foreign Nuon connection subject"},
		"unexpected foreign denial": {foreignError: "invalid_target", wantError: true, wantMessage: "invalid_target: Provider does not exist"},
	} {
		t.Run(name, func(t *testing.T) {
			exchanges := 0
			client := &http.Client{Transport: verificationTransportFunc(func(req *http.Request) (*http.Response, error) {
				status, body := http.StatusOK, ""
				switch req.URL.Host {
				case "sts.googleapis.com":
					require.NoError(t, req.ParseForm())
					parts := strings.Split(req.Form.Get("subject_token"), ".")
					require.Len(t, parts, 3)
					claims, err := base64.RawURLEncoding.DecodeString(parts[1])
					require.NoError(t, err)
					foreign := strings.Contains(string(claims), "org:foreign:connection:foreign")
					switch {
					case foreign && tc.foreignError != "":
						status, body = http.StatusBadRequest, `{"error":"invalid_target","error_description":"Provider does not exist"}`
					case foreign && !tc.allowForeign:
						status, body = http.StatusBadRequest, `{"error":"invalid_grant","error_description":"The given credential is rejected by the attribute condition"}`
					case !foreign && tc.stsError != "" && (tc.failures == 0 || exchanges < tc.failures):
						exchanges++
						status = http.StatusBadRequest
						body = fmt.Sprintf(`{"error":"invalid_grant","error_description":"The given credential is rejected by the attribute condition subject_token=%s","private":"never-include-this"}`, req.Form.Get("subject_token"))
					default:
						body = `{"access_token":"acme-federated-token","token_type":"Bearer","expires_in":3600}`
					}
				case "iamcredentials.googleapis.com":
					if tc.impersonation {
						status, body = http.StatusForbidden, `{"error":{"code":403,"status":"PERMISSION_DENIED","message":"Permission iam.serviceAccounts.getAccessToken denied for acme access_token=acme-secret","details":[{"private":"never-include-this"}]}}`
					} else {
						body = fmt.Sprintf(`{"accessToken":"acme-impersonated-token","expireTime":%q}`, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
					}
				default:
					t.Fatalf("unexpected request: %s", req.URL)
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})}
			connection := &app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", Platform: app.CloudPlatformGCP, Preset: app.CloudConnectionPresetGAR, Principal: "acme@acme-project.iam.gserviceaccount.com", GCPProjectNumber: "123456789012"}
			now, sleeps := time.Now(), 0
			verifier := NewAWSVerifier(issuer, zap.NewNop()).(*awsVerifier)
			verifier.now = func() time.Time { return now }
			verifier.sleep = func(_ context.Context, duration time.Duration) error {
				sleeps++
				now = now.Add(duration)
				return nil
			}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
			result, err := verifier.Verify(ctx, connection, VerifyOptions{RetryIAMPropagation: tc.retry})
			message := result.Message
			if tc.wantError {
				require.Error(t, err)
				message = VerificationErrorMessage(err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantStatus, result.Status)
			}
			require.Equal(t, tc.wantSleeps, sleeps)
			require.Contains(t, message, tc.wantMessage)
			require.NotContains(t, message, "acme-secret")
			require.NotContains(t, message, "never-include-this")
			require.NotContains(t, message, "eyJ")
			if tc.wantStatus == app.CloudConnectionStatusError && !tc.allowForeign {
				require.Contains(t, message, "steps 1 and 2")
			}
		})
	}
}

func TestACRVerification(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		identityCode string
		exchangeBody string
		allowForeign bool
		foreignCode  string
		retry        bool
		failures     int
		wantSleeps   int
		wantStatus   app.CloudConnectionStatus
		wantError    bool
		wantMessage  string
		wantGuidance string
	}{
		"verified":                  {wantStatus: app.CloudConnectionStatusVerified},
		"identity denied":           {identityCode: "AADSTS700016", wantStatus: app.CloudConnectionStatusError, wantMessage: "AADSTS700016: Application not found", wantGuidance: "step 2"},
		"discovery failed":          {identityCode: "AADSTS90061", wantError: true, wantMessage: "AADSTS90061: Application not found", wantGuidance: "publicly reachable issuer discovery"},
		"retry exhausted":           {identityCode: "AADSTS700016", retry: true, wantSleeps: 12, wantStatus: app.CloudConnectionStatusError, wantMessage: "AADSTS700016: Application not found", wantGuidance: "step 2"},
		"retry succeeds":            {identityCode: "AADSTS700016", retry: true, failures: 1, wantSleeps: 1, wantStatus: app.CloudConnectionStatusVerified},
		"exchange denied":           {exchangeBody: `{"errors":[{"code":"DENIED","message":"client has no registry access","detail":{"private":"never-include-this"}}]}`, wantError: true, wantMessage: "DENIED: client has no registry access", wantGuidance: "registry login server"},
		"foreign accepted":          {allowForeign: true, wantStatus: app.CloudConnectionStatusError, wantMessage: "accepts a foreign Nuon connection subject"},
		"unexpected foreign denial": {foreignCode: "AADSTS90061", wantError: true, wantMessage: "AADSTS90061: foreign credential rejected", wantGuidance: "publicly reachable issuer discovery"},
	} {
		t.Run(name, func(t *testing.T) {
			const tenant = "11111111-1111-1111-1111-111111111111"
			accessToken := "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"tid":"`+tenant+`"}`)) + ".c2lnbmF0dXJl"
			attempts := 0
			previous := http.DefaultTransport
			http.DefaultTransport = verificationTransportFunc(func(req *http.Request) (*http.Response, error) {
				status, body := http.StatusOK, ""
				switch {
				case req.URL.Host == "acmeregistry.azurecr.io" && req.URL.Path == "/oauth2/exchange":
					require.NoError(t, req.ParseForm())
					require.Equal(t, accessToken, req.Form.Get("access_token"))
					body = `{"refresh_token":"acme-refresh-token"}`
					if tc.exchangeBody != "" {
						status, body = http.StatusForbidden, tc.exchangeBody
					}
				case req.URL.Host != "login.microsoftonline.com":
					t.Fatalf("unexpected request: %s", req.URL)
				case strings.HasSuffix(req.URL.Path, "/instance"):
					body = `{"tenant_discovery_endpoint":"https://login.microsoftonline.com/` + tenant + `/v2.0/.well-known/openid-configuration","metadata":[{"preferred_network":"login.microsoftonline.com","preferred_cache":"login.microsoftonline.com","aliases":["login.microsoftonline.com"]}]}`
				case strings.HasSuffix(req.URL.Path, "/openid-configuration"):
					body = `{"token_endpoint":"https://login.microsoftonline.com/` + tenant + `/oauth2/v2.0/token","authorization_endpoint":"https://login.microsoftonline.com/` + tenant + `/oauth2/v2.0/authorize","issuer":"https://login.microsoftonline.com/` + tenant + `/v2.0"}`
				case strings.HasSuffix(req.URL.Path, "/token"):
					require.NoError(t, req.ParseForm())
					parts := strings.Split(req.Form.Get("client_assertion"), ".")
					require.Len(t, parts, 3)
					claims, err := base64.RawURLEncoding.DecodeString(parts[1])
					require.NoError(t, err)
					foreign := strings.Contains(string(claims), "org:foreign:connection:foreign")
					code, reason := tc.identityCode, "Application not found"
					if foreign {
						code, reason = tc.foreignCode, "foreign credential rejected"
						if code == "" && !tc.allowForeign {
							code = "AADSTS700213"
						}
					} else {
						attempts++
						if tc.failures > 0 && attempts > tc.failures {
							code = ""
						}
					}
					body = fmt.Sprintf(`{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, accessToken)
					if code != "" {
						status = http.StatusBadRequest
						encoded, err := json.Marshal(map[string]string{"error": "invalid_client", "error_description": code + ": " + reason + " client_assertion=" + req.Form.Get("client_assertion"), "private": "never-include-this"})
						require.NoError(t, err)
						body = string(encoded)
					}
				default:
					t.Fatalf("unexpected request: %s", req.URL)
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})
			t.Cleanup(func() { http.DefaultTransport = previous })
			connection := &app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", Platform: app.CloudPlatformAzure, Preset: app.CloudConnectionPresetACR, TargetID: tenant, Principal: "22222222-2222-2222-2222-222222222222", AzureTenantID: tenant, AzureRegistryID: "/subscriptions/" + tenant + "/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry", AzureRegistryLoginServer: "acmeregistry.azurecr.io"}
			now, sleeps := time.Now(), 0
			verifier := NewAWSVerifier(issuer, zap.NewNop()).(*awsVerifier)
			verifier.now = func() time.Time { return now }
			verifier.sleep = func(_ context.Context, duration time.Duration) error {
				sleeps++
				now = now.Add(duration)
				return nil
			}
			result, err := verifier.Verify(context.Background(), connection, VerifyOptions{RetryIAMPropagation: tc.retry})
			message := result.Message
			if tc.wantError {
				require.Error(t, err)
				message = VerificationErrorMessage(err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantStatus, result.Status)
			}
			require.Equal(t, tc.wantSleeps, sleeps)
			require.Contains(t, message, tc.wantMessage)
			require.Contains(t, message, tc.wantGuidance)
			require.NotContains(t, message, "eyJ")
			require.NotContains(t, message, "never-include-this")
		})
	}
}
