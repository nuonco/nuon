package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func TestAWSVerification(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
	require.NoError(t, err)
	for _, tc := range []struct {
		name         string
		preset       app.CloudConnectionPreset
		identityOnly bool
		allowForeign bool
		denyStacks   bool
		wantStatus   app.CloudConnectionStatus
		wantProbes   int
	}{
		{"stacks read probe", app.CloudConnectionPresetStacks, false, false, false, app.CloudConnectionStatusVerified, 1},
		{"stacks denied", app.CloudConnectionPresetStacks, false, false, true, app.CloudConnectionStatusError, 1},
		{"custom ignores stack permissions", app.CloudConnectionPresetCustom, false, false, true, app.CloudConnectionStatusVerified, 0},
		{"cron identity only", app.CloudConnectionPresetStacks, true, false, true, app.CloudConnectionStatusVerified, 0},
		{"foreign subject accepted", app.CloudConnectionPresetCustom, false, true, false, app.CloudConnectionStatusError, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probes, exchanges := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				w.Header().Set("Content-Type", "text/xml")
				deny := func() {
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprint(w, `<ErrorResponse><Error><Code>AccessDenied</Code><Message>Denied</Message></Error></ErrorResponse>`)
				}
				switch r.Form.Get("Action") {
				case "AssumeRoleWithWebIdentity":
					exchanges++
					parts := strings.Split(r.Form.Get("WebIdentityToken"), ".")
					if len(parts) != 3 {
						t.Error("expected signed OIDC token")
						deny()
						return
					}
					claims, _ := base64.RawURLEncoding.DecodeString(parts[1])
					if strings.Contains(string(claims), "org:foreign:connection:foreign") && !tc.allowForeign {
						deny()
						return
					}
					fmt.Fprint(w, `<AssumeRoleWithWebIdentityResponse><AssumeRoleWithWebIdentityResult><Credentials><AccessKeyId>acme</AccessKeyId><SecretAccessKey>acme-secret</SecretAccessKey><SessionToken>acme-token</SessionToken></Credentials></AssumeRoleWithWebIdentityResult></AssumeRoleWithWebIdentityResponse>`)
				case "GetCallerIdentity":
					fmt.Fprint(w, `<GetCallerIdentityResponse><GetCallerIdentityResult><Account>123456789012</Account><Arn>arn:aws:sts::123456789012:assumed-role/acme/session</Arn></GetCallerIdentityResult></GetCallerIdentityResponse>`)
				case "DescribeStacks":
					probes++
					if tc.denyStacks {
						deny()
						return
					}
					fmt.Fprint(w, `<DescribeStacksResponse><DescribeStacksResult><Stacks/></DescribeStacksResult></DescribeStacksResponse>`)
				default:
					t.Errorf("unexpected action %q", r.Form.Get("Action"))
					deny()
				}
			}))
			defer server.Close()
			t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)
			t.Setenv("AWS_ENDPOINT_URL_CLOUDFORMATION", server.URL)
			t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
			connection := &app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", Platform: app.CloudPlatformAWS, Preset: tc.preset, TargetID: "123456789012", Principal: "arn:aws:iam::123456789012:role/acme", DefaultRegion: "us-east-1"}
			result, err := NewAWSVerifier(issuer).Verify(context.Background(), connection, VerifyOptions{IdentityOnly: tc.identityOnly})
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, result.Status)
			require.Equal(t, 2, exchanges)
			require.Equal(t, tc.wantProbes, probes)
		})
	}
}
