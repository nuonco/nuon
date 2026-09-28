package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func TestAWSVerification(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_DEFAULT_PROFILE", "")
	t.Setenv("AWS_CONFIG_FILE", os.DevNull)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", os.DevNull)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		preset         app.CloudConnectionPreset
		identityOnly   bool
		allowForeign   bool
		denyStacks     bool
		assumeError    string
		wantStatus     app.CloudConnectionStatus
		wantMessage    string
		wantProbes     int
		retry          bool
		assumeFailures int
		wantExchanges  int
		wantSleeps     int
	}{
		"denied then allowed": {
			preset: app.CloudConnectionPresetStacks, retry: true, assumeError: "AccessDenied", assumeFailures: 1,
			wantStatus: app.CloudConnectionStatusVerified, wantProbes: 1, wantExchanges: 3, wantSleeps: 1,
		},
		"denied past deadline": {
			preset: app.CloudConnectionPresetStacks, retry: true, assumeError: "AccessDenied",
			wantStatus: app.CloudConnectionStatusError, wantExchanges: 12, wantSleeps: 12,
			wantMessage: "Nuon OIDC identity is not trusted by this role.",
		},
		"stacks read probe": {
			preset: app.CloudConnectionPresetStacks, wantStatus: app.CloudConnectionStatusVerified, wantProbes: 1,
		},
		"stacks denied": {
			preset: app.CloudConnectionPresetStacks, retry: true, denyStacks: true, wantStatus: app.CloudConnectionStatusError, wantProbes: 1,
			wantMessage: "The role lacks CloudFormation read access required to manage install stacks.",
		},
		"invalid identity token": {
			preset: app.CloudConnectionPresetStacks, retry: true, assumeError: "InvalidIdentityToken", wantStatus: app.CloudConnectionStatusError,
			wantMessage: "AWS could not validate Nuon's identity token. Create the OIDC provider for this issuer first (step 1).",
		},
		"role not trusted": {
			preset: app.CloudConnectionPresetStacks, assumeError: "AccessDenied", wantStatus: app.CloudConnectionStatusError,
			wantMessage: "Nuon OIDC identity is not trusted by this role.",
		},
		"custom ignores stack permissions": {
			preset: app.CloudConnectionPresetCustom, denyStacks: true, wantStatus: app.CloudConnectionStatusVerified,
		},
		"cron identity only": {
			preset: app.CloudConnectionPresetStacks, identityOnly: true, denyStacks: true, wantStatus: app.CloudConnectionStatusVerified,
		},
		"foreign subject accepted": {
			preset: app.CloudConnectionPresetCustom, retry: true, allowForeign: true, wantStatus: app.CloudConnectionStatusError,
		},
	} {
		t.Run(name, func(t *testing.T) {
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
					if tc.assumeError != "" && (tc.assumeFailures == 0 || exchanges <= tc.assumeFailures) {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprintf(w, `<ErrorResponse><Error><Code>%s</Code><Message>Raw AWS diagnostic</Message></Error></ErrorResponse>`, tc.assumeError)
						return
					}
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
			now, sleeps := time.Now(), 0
			verifier := NewAWSVerifier(issuer).(*awsVerifier)
			verifier.now = func() time.Time { return now }
			verifier.sleep = func(ctx context.Context, duration time.Duration) error {
				require.Equal(t, 5*time.Second, duration)
				sleeps++
				now = now.Add(duration)
				return nil
			}
			result, err := verifier.Verify(context.Background(), connection, VerifyOptions{IdentityOnly: tc.identityOnly, RetryIAMPropagation: tc.retry})
			require.NoError(t, err)
			require.Equal(t, tc.wantSleeps, sleeps)
			require.Equal(t, tc.wantStatus, result.Status)
			if tc.wantMessage != "" {
				require.Equal(t, tc.wantMessage, result.Message)
			}
			if tc.wantExchanges != 0 {
				require.Equal(t, tc.wantExchanges, exchanges)
			} else if tc.assumeError != "" {
				require.Equal(t, 1, exchanges)
			} else {
				require.Equal(t, 2, exchanges)
			}
			require.Equal(t, tc.wantProbes, probes)
		})
	}
}

func TestVerificationErrorMessage(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want string
	}{
		"AWS error": {
			err:  fmt.Errorf("probe: %w", &smithy.GenericAPIError{Code: "Throttling", Message: "Raw AWS diagnostic"}),
			want: "Verification failed: Throttling",
		},
		"timeout": {
			err:  fmt.Errorf("probe: %w", context.DeadlineExceeded),
			want: "Verification failed: request timed out",
		},
		"canceled": {
			err:  context.Canceled,
			want: "Verification failed: request canceled",
		},
		"unexpected error": {
			err:  fmt.Errorf("a long internal diagnostic"),
			want: "Verification failed: unable to complete the AWS verification request",
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, cloudconnections.VerificationErrorMessage(tc.err))
		})
	}
}
