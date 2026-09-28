package activities

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
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
	}{
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
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			connection := app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", CreatedByID: "acct_acme", Name: "acme", Principal: "arn:aws:iam::123456789012:role/acme", Status: tc.status, Platform: app.CloudPlatformAWS}
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
				require.Equal(t, "result from AWS", update.StatusMessage)
				require.WithinDuration(t, time.Now(), *update.LastVerifiedAt, time.Second)
				if tc.result == app.CloudConnectionStatusVerified {
					require.Equal(t, app.CloudConnectionAuthModeOIDC, update.AuthMode)
				}
				tx.RowsAffected = 1
			}))
			a := &Activities{db: db, l: zap.NewNop(), verifier: fakeVerifier(func(_ context.Context, c *app.CloudConnection, opts cloudconnections.VerifyOptions) (cloudconnections.VerificationResult, error) {
				calls++
				require.Equal(t, connection, *c)
				require.Equal(t, cloudconnections.VerifyOptions{IdentityOnly: !tc.onDemand, RetryIAMPropagation: tc.wantRetry}, opts)
				return cloudconnections.VerificationResult{Status: tc.result, Message: "result from AWS"}, nil
			}), enqueueOrgSignal: func(ctx context.Context, params orgshelpers.EnqueueOrgSignalParams) error {
				signals++
				require.Equal(t, connection.CreatedByID, ctx.Value(keys.AccountIDCtxKey))
				require.Equal(t, "result from AWS", params.Signal.(*verificationfailed.Signal).Message)
				return nil
			}}
			require.NoError(t, a.Reverify(context.Background(), ReverifyRequest{CloudConnectionID: connection.ID, OnDemand: tc.onDemand}))
			require.Equal(t, tc.wantUpdates, updates)
			require.Equal(t, tc.wantSignals, signals)
			if tc.deleted {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}
