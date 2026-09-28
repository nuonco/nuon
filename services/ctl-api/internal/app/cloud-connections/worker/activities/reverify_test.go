package activities

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type fakeVerifier func(context.Context, *app.CloudConnection, cloudconnections.VerifyOptions) (cloudconnections.VerificationResult, error)

func (f fakeVerifier) Verify(ctx context.Context, c *app.CloudConnection, opts cloudconnections.VerifyOptions) (cloudconnections.VerificationResult, error) {
	return f(ctx, c, opts)
}

func TestReverify(t *testing.T) {
	for name, tc := range map[string]struct {
		status            app.CloudConnectionStatus
		result            app.CloudConnectionStatus
		deleted           bool
		onDemand          bool
		previouslyChecked bool
		wantRetry         bool
		wantUpdates       int
		wantSignals       int
		awsPreset         app.CloudConnectionPreset
		wantProbes        int
		message           string
	}{
		"periodic stacks permissions denied": {
			status: app.CloudConnectionStatusVerified, result: app.CloudConnectionStatusError,
			awsPreset: app.CloudConnectionPresetStacks, wantProbes: 1, wantUpdates: 1, wantSignals: 1,
			message: "The role lacks CloudFormation read access required to manage install stacks.",
		},
		"periodic custom skips stack permissions": {
			status: app.CloudConnectionStatusVerified, result: app.CloudConnectionStatusVerified,
			awsPreset: app.CloudConnectionPresetCustom, wantUpdates: 1, message: "Cloud connection verified.",
		},
		"verified succeeds":          {status: app.CloudConnectionStatusVerified, result: app.CloudConnectionStatusVerified, wantUpdates: 1},
		"verified fails":             {status: app.CloudConnectionStatusVerified, result: app.CloudConnectionStatusError, wantUpdates: 1, wantSignals: 1},
		"pending succeeds":           {status: app.CloudConnectionStatusPending, result: app.CloudConnectionStatusVerified, wantUpdates: 1},
		"pending fails":              {status: app.CloudConnectionStatusPending, result: app.CloudConnectionStatusError},
		"pending on demand succeeds": {status: app.CloudConnectionStatusPending, result: app.CloudConnectionStatusVerified, onDemand: true, wantRetry: true, wantUpdates: 1},
		"pending on demand fails":    {status: app.CloudConnectionStatusPending, result: app.CloudConnectionStatusError, onDemand: true, wantRetry: true, wantUpdates: 1, wantSignals: 1},
		"verified on demand fails":   {status: app.CloudConnectionStatusVerified, result: app.CloudConnectionStatusError, onDemand: true, previouslyChecked: true, wantUpdates: 1, wantSignals: 1},
		"previous failure on demand": {status: app.CloudConnectionStatusError, result: app.CloudConnectionStatusError, onDemand: true, previouslyChecked: true, wantUpdates: 1, wantSignals: 1},
		"deleted":                    {deleted: true},
	} {
		t.Run(name, func(t *testing.T) {
			message := tc.message
			if message == "" {
				message = "result from AWS"
			}
			probes := 0
			var awsVerifier cloudconnections.Verifier
			if tc.awsPreset != "" {
				t.Setenv("AWS_PROFILE", "")
				t.Setenv("AWS_DEFAULT_PROFILE", "")
				t.Setenv("AWS_CONFIG_FILE", os.DevNull)
				t.Setenv("AWS_SHARED_CREDENTIALS_FILE", os.DevNull)
				t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, r.ParseForm())
					w.Header().Set("Content-Type", "text/xml")
					switch r.Form.Get("Action") {
					case "AssumeRoleWithWebIdentity":
						if r.Form.Get("RoleSessionName") != "nuon-cloud-connection-negative-probe" {
							fmt.Fprint(w, `<AssumeRoleWithWebIdentityResponse><AssumeRoleWithWebIdentityResult><Credentials><AccessKeyId>acme</AccessKeyId><SecretAccessKey>acme-secret</SecretAccessKey><SessionToken>acme-token</SessionToken></Credentials></AssumeRoleWithWebIdentityResult></AssumeRoleWithWebIdentityResponse>`)
							return
						}
					case "GetCallerIdentity":
						fmt.Fprint(w, `<GetCallerIdentityResponse><GetCallerIdentityResult><Account>123456789012</Account><Arn>arn:aws:sts::123456789012:assumed-role/acme/session</Arn></GetCallerIdentityResult></GetCallerIdentityResponse>`)
						return
					case "DescribeStacks":
						probes++
					default:
						t.Errorf("unexpected AWS action %q", r.Form.Get("Action"))
					}
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprint(w, `<ErrorResponse><Error><Code>AccessDenied</Code><Message>Denied</Message></Error></ErrorResponse>`)
				}))
				defer server.Close()
				t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)
				t.Setenv("AWS_ENDPOINT_URL_CLOUDFORMATION", server.URL)
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)
				issuer, err := oidcissuer.New("https://api.example.com", key, "acme")
				require.NoError(t, err)
				awsVerifier = cloudconnections.NewAWSVerifier(issuer)
			}
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			connection := app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", CreatedByID: "acct_acme", Name: "acme", Principal: "arn:aws:iam::123456789012:role/acme", Status: tc.status, Platform: app.CloudPlatformAWS, Preset: tc.awsPreset, TargetID: "123456789012", DefaultRegion: "us-east-1"}
			if tc.previouslyChecked || tc.status == app.CloudConnectionStatusVerified {
				checked := time.Now().Add(-time.Hour)
				connection.LastVerifiedAt = &checked
			}
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("load", func(tx *gorm.DB) {
				require.Contains(t, tx.Statement.SQL.String(), `"cloud_connections"."deleted_at"`)
				if tc.deleted {
					tx.AddError(gorm.ErrRecordNotFound)
					return
				}
				*tx.Statement.Dest.(*app.CloudConnection) = connection
			}))
			updates, signals, calls := 0, 0, 0
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("save", func(tx *gorm.DB) {
				updates++
				update := tx.Statement.Dest.(app.CloudConnection)
				require.Equal(t, tc.result, update.Status)
				require.Equal(t, message, update.StatusMessage)
				require.WithinDuration(t, time.Now(), *update.LastVerifiedAt, time.Second)
				if tc.result == app.CloudConnectionStatusVerified {
					require.Equal(t, app.CloudConnectionAuthModeOIDC, update.AuthMode)
				}
				tx.RowsAffected = 1
			}))
			a := &Activities{db: db, l: zap.NewNop(), verifier: fakeVerifier(func(ctx context.Context, c *app.CloudConnection, opts cloudconnections.VerifyOptions) (cloudconnections.VerificationResult, error) {
				calls++
				require.Equal(t, connection, *c)
				require.Equal(t, cloudconnections.VerifyOptions{RetryIAMPropagation: tc.wantRetry}, opts)
				if awsVerifier != nil {
					return awsVerifier.Verify(ctx, c, opts)
				}
				return cloudconnections.VerificationResult{Status: tc.result, Message: message}, nil
			}), enqueueOrgSignal: func(ctx context.Context, params orgshelpers.EnqueueOrgSignalParams) error {
				signals++
				require.Equal(t, connection.CreatedByID, ctx.Value(keys.AccountIDCtxKey))
				require.Equal(t, message, params.Signal.(*verificationfailed.Signal).Message)
				return nil
			}}
			require.NoError(t, a.Reverify(context.Background(), ReverifyRequest{CloudConnectionID: connection.ID, OnDemand: tc.onDemand}))
			require.Equal(t, tc.wantUpdates, updates)
			require.Equal(t, tc.wantSignals, signals)
			require.Equal(t, tc.wantProbes, probes)
			if tc.deleted {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}
