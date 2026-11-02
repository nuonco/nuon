package helpers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type imageAuthTransport func(*http.Request) (*http.Response, error)

func (f imageAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestPinnedCloudImageAuthorization(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
	require.NoError(t, err)
	for name, tc := range map[string]struct{ denied bool }{"success": {}, "cloud rejects token": {denied: true}} {
		for _, platform := range []string{"ecr", "gar"} {
			t.Run(platform+"/"+name, func(t *testing.T) {
				db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
				require.NoError(t, err)
				require.NoError(t, db.Callback().Query().After("gorm:query").Register("image-cloud", func(tx *gorm.DB) {
					require.Contains(t, tx.Statement.Vars, "cc_saved")
					require.Contains(t, tx.Statement.Vars, "org_acme")
					connection := tx.Statement.Dest.(*app.CloudConnection)
					*connection = app.CloudConnection{ID: "cc_saved", OrgID: "org_acme", Platform: app.CloudPlatformAWS, Status: app.CloudConnectionStatusVerified, DefaultRegion: "us-east-1", Principal: "arn:aws:iam::123456789012:role/acme", GCPProjectNumber: "123456789012"}
					tx.RowsAffected = 1
				}))
				calls := 0
				respond := func(req *http.Request) (*http.Response, error) {
					calls++
					status, body, contentType := http.StatusOK, "", "application/json"
					if platform == "ecr" {
						if req.Header.Get("X-Amz-Target") != "" {
							body = fmt.Sprintf(`{"authorizationData":[{"authorizationToken":%q,"proxyEndpoint":"https://123456789012.dkr.ecr.us-east-1.amazonaws.com"}]}`, base64.StdEncoding.EncodeToString([]byte("AWS:acme-token")))
						} else {
							require.NoError(t, req.ParseForm())
							require.Equal(t, "AssumeRoleWithWebIdentity", req.Form.Get("Action"))
							require.Equal(t, "arn:aws:iam::123456789012:role/acme", req.Form.Get("RoleArn"))
							contentType = "text/xml"
							body = `<AssumeRoleWithWebIdentityResponse><AssumeRoleWithWebIdentityResult><Credentials><AccessKeyId>acme-key</AccessKeyId><SecretAccessKey>acme-secret</SecretAccessKey><SessionToken>acme-session</SessionToken><Expiration>` + time.Now().Add(time.Hour).UTC().Format(time.RFC3339) + `</Expiration></Credentials></AssumeRoleWithWebIdentityResult></AssumeRoleWithWebIdentityResponse>`
							if tc.denied {
								status = http.StatusBadRequest
								body = `<ErrorResponse><Error><Code>InvalidIdentityToken</Code><Message>acme provider rejects token</Message></Error></ErrorResponse>`
							}
						}
					} else {
						switch req.URL.Host {
						case "sts.googleapis.com":
							body = `{"access_token":"acme-federated","token_type":"Bearer","expires_in":3600}`
							if tc.denied {
								status, body = http.StatusBadRequest, `{"error":"invalid_grant","error_description":"acme provider rejects token"}`
							}
						case "iamcredentials.googleapis.com":
							body = fmt.Sprintf(`{"accessToken":"acme-token","expireTime":%q}`, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
						default:
							t.Fatalf("unexpected cloud host: %s", req.URL.Host)
						}
					}
					return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				}
				ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: imageAuthTransport(respond)})
				acts := &mockImageAuthActivities{}
				h := &Helpers{db: db, issuer: issuer, imageAuth: acts}
				var token string
				if platform == "ecr" {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						response, err := respond(req)
						require.NoError(t, err)
						w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
						w.WriteHeader(response.StatusCode)
						_, _ = io.Copy(w, response.Body)
					}))
					defer server.Close()
					t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)
					t.Setenv("AWS_ENDPOINT_URL_ECR", server.URL)
					t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
					auth, authErr := h.ECRAuthorization(ctx, "org_acme", &app.AWSECRImageConfig{ID: "saved", CloudConnectionID: "cc_saved", AWSRegion: "us-east-1"}, "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme")
					err = authErr
					if auth != nil {
						token = auth.RegistryToken
					}
				} else {
					auth, authErr := h.GARAuthorization(ctx, "org_acme", &app.GCPGARImageConfig{ID: "saved", CloudConnectionID: "cc_saved", GCPRegion: "us-central1", GCPProjectID: "acme"}, "us-central1-docker.pkg.dev/acme/repo/image")
					err = authErr
					if auth != nil {
						token = auth.AccessToken
					}
				}
				if tc.denied {
					require.ErrorContains(t, err, "acme provider rejects token")
					require.Empty(t, token)
				} else {
					require.NoError(t, err)
					require.Equal(t, "acme-token", token)
				}
				require.Positive(t, calls)
				acts.AssertNotCalled(t, "GetECRAccessToken", mock.Anything)
				acts.AssertNotCalled(t, "GetGARAccessToken", mock.Anything)
			})
		}
	}
}
