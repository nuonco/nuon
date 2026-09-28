package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type fakeSTS struct{ err error }

func (s fakeSTS) AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	return nil, s.err
}

func TestAssumeRoleFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		err         error
		wantMessage string
	}{
		"access denied": {
			err:         &smithy.GenericAPIError{Code: "AccessDenied", Message: "raw diagnostic"},
			wantMessage: "Nuon OIDC identity is not trusted by this role.",
		},
		"invalid identity token": {
			err:         &smithy.GenericAPIError{Code: "InvalidIdentityToken", Message: "raw diagnostic"},
			wantMessage: "AWS could not validate Nuon's identity token. Create the OIDC provider for this issuer first (step 1).",
		},
		"other AWS error": {
			err:         &smithy.GenericAPIError{Code: "Throttling", Message: "raw diagnostic"},
			wantMessage: "Verification failed: Throttling",
		},
		"network error": {err: errors.New("connection reset by peer")},
		"timeout":       {err: context.DeadlineExceeded},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{
				DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			previous := time.Now().Add(-time.Hour)
			connection := &app.CloudConnection{
				ID: "cc_acme", OrgID: "org_acme", CreatedByID: "acct_acme", Name: "acme",
				Platform: app.CloudPlatformAWS, Principal: "arn:aws:iam::123456789012:role/acme",
				Status: app.CloudConnectionStatusVerified, StatusMessage: "Cloud connection verified.", LastVerifiedAt: &previous,
			}
			stored := *connection
			updates, signals := 0, 0
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("capture", func(tx *gorm.DB) {
				updates++
				update := tx.Statement.Dest.(app.CloudConnection)
				require.Equal(t, []string{"status", "status_message", "last_verified_at"}, tx.Statement.Selects)
				require.Contains(t, tx.Statement.SQL.String(), `"cloud_connections"."org_id"`)
				require.Contains(t, tx.Statement.SQL.String(), `"cloud_connections"."principal"`)
				require.Contains(t, tx.Statement.Vars, connection.OrgID)
				require.Contains(t, tx.Statement.Vars, connection.ID)
				require.Contains(t, tx.Statement.Vars, connection.Principal)
				stored.Status, stored.StatusMessage, stored.LastVerifiedAt = update.Status, update.StatusMessage, update.LastVerifiedAt
				tx.RowsAffected = 1
			}))
			h := &Helpers{db: db, enqueueOrgSignal: func(ctx context.Context, params orgshelpers.EnqueueOrgSignalParams) error {
				signals++
				require.Equal(t, connection.CreatedByID, ctx.Value(keys.AccountIDCtxKey))
				require.Equal(t, connection.OrgID, params.OrgID)
				require.Equal(t, &verificationfailed.Signal{
					ConnectionID: connection.ID, ConnectionName: connection.Name, OrgID: connection.OrgID,
					Platform: connection.Platform, Message: tc.wantMessage,
				}, params.Signal)
				require.Contains(t, params.IdempotencyKey, connection.ID)
				return nil
			}}
			_, err = h.assumeRole(context.Background(), connection, fakeSTS{err: tc.err}, &sts.AssumeRoleWithWebIdentityInput{})
			require.ErrorIs(t, err, tc.err)
			if tc.wantMessage == "" {
				require.Zero(t, updates)
				require.Zero(t, signals)
				require.Equal(t, *connection, stored)
			} else {
				require.Equal(t, 1, updates)
				require.Equal(t, 1, signals)
				require.Equal(t, app.CloudConnectionStatusError, stored.Status)
				require.Equal(t, tc.wantMessage, stored.StatusMessage)
				require.WithinDuration(t, time.Now(), *stored.LastVerifiedAt, time.Second)
			}
		})
	}
}
