package cloudconnections

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestVerificationDiagnostics(t *testing.T) {
	for name, tc := range map[string]struct {
		err        error
		assumeRole bool
		want       string
		guidance   string
	}{
		"AWS denial": {
			err: &smithy.GenericAPIError{Code: "AccessDenied", Message: "Role acme denies sts:AssumeRoleWithWebIdentity token=acme-secret"}, assumeRole: true,
			want: "AccessDenied: Role acme denies sts:AssumeRoleWithWebIdentity token=[REDACTED]", guidance: "Nuon OIDC identity is not trusted by this role.",
		},
		"AWS invalid identity": {
			err: fmt.Errorf("assume role: %w", &smithy.GenericAPIError{Code: "InvalidIdentityToken", Message: "No OpenIDConnect provider found for issuer https://api.example.com"}), assumeRole: true,
			want: "InvalidIdentityToken: No OpenIDConnect provider found", guidance: "Create the OIDC provider for this issuer first (step 1).",
		},
		"generic AWS failure": {
			err: &smithy.GenericAPIError{Code: "Throttling", Message: "Rate limit exceeded client_secret=acme-secret"}, assumeRole: true,
			want: "Throttling: Rate limit exceeded client_secret=[REDACTED]",
		},
		"OAuth failure": {
			err:  &oauth2.RetrieveError{ErrorCode: "invalid_grant", ErrorDescription: "Credential is rejected access_token=acme-secret", Body: []byte("never-include-this")},
			want: "invalid_grant: Credential is rejected access_token=[REDACTED]",
		},
		"OAuth unknown response": {
			err:  &oauth2.RetrieveError{Response: &http.Response{StatusCode: http.StatusBadGateway}, Body: []byte("never-include-this")},
			want: "OAuth token request returned status 502",
		},
		"Google unknown response": {
			err: fmt.Errorf("oauth2/google: status code 502: never-include-this"), want: "status code 502",
		},
		"Google unknown JSON": {
			err: fmt.Errorf(`oauth2/google: status code 403: {"error":{"details":"never-include-this"},"access_token":"never-include-this"}`), want: "status code 403",
		},
		"Google oversized response": {
			err: fmt.Errorf(`oauth2/google: status code 403: {"error":"invalid_grant","error_description":"%snever-include-this"}`, strings.Repeat("x", 32*1024)), want: "status code 403",
		},
		"Entra unknown response": {
			err:  &azidentity.AuthenticationFailedError{RawResponse: &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("never-include-this"))}},
			want: "Microsoft Entra returned status 502",
		},
		"Entra OAuth details": {
			err:  &azidentity.AuthenticationFailedError{RawResponse: &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{"error":"invalid_client","error_description":"AADSTS700016: Application not found client_assertion=acme-secret","private":"never-include-this"}`))}},
			want: "invalid_client: AADSTS700016: Application not found client_assertion=[REDACTED]",
		},
		"registry context and signed URL": {
			err:  fmt.Errorf("fetch image acme: MANIFEST_UNKNOWN at https://user:acme-secret@registry.example.com/acme?X-Amz-Signature=acme-secret"),
			want: "fetch image acme: MANIFEST_UNKNOWN at https://[REDACTED]@registry.example.com/acme?X-Amz-Signature=[REDACTED]",
		},
		"assertion and bearer": {
			err:  fmt.Errorf("invalid_grant subject_token=acme-secret client_assertion=acme-secret Bearer acme-secret eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhY21lIn0.c2ln"),
			want: "invalid_grant subject_token=[REDACTED] client_assertion=[REDACTED] Bearer [REDACTED] [REDACTED]",
		},
		"redact before truncation": {
			err: fmt.Errorf("DENIED access_token=%s", strings.Repeat("acme-secret", 1000)), want: "DENIED access_token=[REDACTED]",
		},
		"bounded unicode": {
			err: fmt.Errorf("DENIED %s", strings.Repeat("理由", 1000)), want: "DENIED 理由",
		},
		"missing cause": {want: "unable to complete the cloud verification request"},
	} {
		t.Run(name, func(t *testing.T) {
			message := VerificationErrorMessage(tc.err)
			if tc.assumeRole {
				message = AssumeRoleErrorMessage(tc.err)
			}
			require.Contains(t, message, tc.want)
			require.Contains(t, message, tc.guidance)
			require.NotContains(t, message, "acme-secret")
			require.NotContains(t, message, "never-include-this")
			require.LessOrEqual(t, len(message), 2300)
			require.True(t, utf8.ValidString(message))
		})
	}
}
