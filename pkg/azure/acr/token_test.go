package acr

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/stretchr/testify/require"
)

type tokenCredentialFunc func(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error)

func (f tokenCredentialFunc) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return f(ctx, opts)
}

type exchangeTransportFunc func(*http.Request) (*http.Response, error)

func (f exchangeTransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type exchangeBody struct {
	io.Reader
	read   int
	closed bool
}

func (b *exchangeBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}

func (b *exchangeBody) Close() error {
	b.closed = true
	return nil
}

func TestGetRepositoryTokenWithCredential(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"tid":"acme-tenant"}`)) + ".c2lnbmF0dXJl"
	for name, tc := range map[string]struct {
		status       int
		body         string
		want         string
		wantToken    string
		wantRedacted bool
	}{
		"success": {status: http.StatusOK, body: `{"refresh_token":"acme-refresh"}`, wantToken: "acme-refresh"},
		"oauth denial": {
			status: http.StatusUnauthorized, body: `{"error":"invalid_grant","error_description":"AADSTS700016: Application not found in tenant","access_token":"never-include-this"}`,
			want: "invalid_grant: AADSTS700016: Application not found in tenant",
		},
		"registry denial": {
			status: http.StatusForbidden, body: `{"errors":[{"code":"DENIED","message":"client has no access to registry","detail":{"secret":"never-include-this"}},{"code":"UNAUTHORIZED","message":"authentication required"}]}`,
			want: "DENIED: client has no access to registry; UNAUTHORIZED: authentication required",
		},
		"echoed request credential": {
			status: http.StatusUnauthorized, body: fmt.Sprintf(`{"error":"invalid_token","error_description":"Rejected %s access_token=unrelated-secret"}`, token),
			want: "invalid_token: Rejected [REDACTED] access_token=[REDACTED]", wantRedacted: true,
		},
		"malformed body":           {status: http.StatusBadGateway, body: `private arbitrary response never-include-this`, want: "returned status 502"},
		"unknown json":             {status: http.StatusBadRequest, body: `{"message":"never-include-this","refresh_token":"never-include-this"}`, want: "returned status 400"},
		"invalid schema":           {status: http.StatusBadRequest, body: `{"error":{"secret":"never-include-this"}}`, want: "returned status 400"},
		"oversized body":           {status: http.StatusBadGateway, body: strings.Repeat("x", 1<<20), want: "returned status 502"},
		"bounded diagnostic":       {status: http.StatusForbidden, body: `{"error":"DENIED","error_description":"` + strings.Repeat("reason ", 1000) + `"}`, want: "DENIED: reason"},
		"redirect is not followed": {status: http.StatusTemporaryRedirect, body: `never-include-this`, want: "returned status 307"},
	} {
		t.Run(name, func(t *testing.T) {
			body := &exchangeBody{Reader: strings.NewReader(tc.body)}
			previous := exchangeClient.Transport
			exchangeClient.Transport = exchangeTransportFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "https://acmeregistry.azurecr.io/oauth2/exchange", req.URL.String())
				require.NoError(t, req.ParseForm())
				require.Equal(t, token, req.Form.Get("access_token"))
				require.Equal(t, "acme-tenant", req.Form.Get("tenant"))
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Location": {"https://foreign.example.com"}}, Body: body, Request: req}, nil
			})
			t.Cleanup(func() { exchangeClient.Transport = previous })
			credential := tokenCredentialFunc(func(_ context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
				require.Equal(t, []string{RegistryScope}, opts.Scopes)
				return azcore.AccessToken{Token: token, ExpiresOn: time.Now().Add(time.Hour)}, nil
			})
			got, err := GetRepositoryTokenWithCredential(context.Background(), credential, "", "acmeregistry.azurecr.io")
			require.True(t, body.closed)
			require.LessOrEqual(t, body.read, 16*1024+1)
			if tc.wantToken != "" {
				require.NoError(t, err)
				require.Equal(t, tc.wantToken, got)
				return
			}
			require.Empty(t, got)
			require.ErrorContains(t, err, tc.want)
			require.NotContains(t, err.Error(), token)
			require.NotContains(t, err.Error(), "never-include-this")
			require.LessOrEqual(t, len(err.Error()), 2200)
			if tc.wantRedacted {
				require.NotContains(t, err.Error(), "unrelated-secret")
			}
		})
	}
}
