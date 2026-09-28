package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestVerificationInProgress(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)
	svc := &service{db: db}
	now := time.Now().In(time.FixedZone("offset", 19800))
	before, after := now.Add(-time.Nanosecond), now.Add(time.Nanosecond)
	for name, tc := range map[string]struct {
		requested, verified *time.Time
		want                bool
	}{
		"never requested or verified":   {},
		"verified without request":      {verified: &now},
		"requested never verified":      {requested: &now, want: true},
		"requested before verification": {requested: &before, verified: &now},
		"equal timestamps":              {requested: &now, verified: &now},
		"requested after verification":  {requested: &after, verified: &now, want: true},
	} {
		t.Run(name, func(t *testing.T) {
			response, err := svc.response(context.Background(), &app.CloudConnection{VerificationRequestedAt: tc.requested, LastVerifiedAt: tc.verified})
			require.NoError(t, err)
			require.Equal(t, tc.want, response.VerificationInProgress)
			if tc.requested != nil {
				require.Equal(t, time.UTC, response.VerificationRequestedAt.Location())
				require.True(t, tc.requested.Equal(*response.VerificationRequestedAt))
			}
		})
	}
}

func TestVerifyEnqueues(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	recent, boundary, expired := now.Add(-2*time.Minute+time.Microsecond), now.Add(-2*time.Minute), now.Add(-2*time.Minute-time.Microsecond)
	for name, tc := range map[string]struct {
		status       app.CloudConnectionStatus
		enqueueError error
		requested    *time.Time
		completed    bool
		deduplicated bool
	}{
		"pending":                  {status: app.CloudConnectionStatusPending},
		"verified":                 {status: app.CloudConnectionStatusVerified},
		"error":                    {status: app.CloudConnectionStatusError},
		"enqueue fails":            {status: app.CloudConnectionStatusVerified, enqueueError: errors.New("queue full")},
		"pending recent request":   {status: app.CloudConnectionStatusPending, requested: &recent, deduplicated: true},
		"verified recent request":  {status: app.CloudConnectionStatusVerified, requested: &recent, deduplicated: true},
		"at two minute boundary":   {status: app.CloudConnectionStatusPending, requested: &boundary},
		"past two minute boundary": {status: app.CloudConnectionStatusPending, requested: &expired},
		"recent completed request": {status: app.CloudConnectionStatusVerified, requested: &recent, completed: true},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			lastVerified := now.Add(-time.Hour)
			if tc.completed {
				lastVerified = now.Add(-time.Second)
			}
			connection := app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", Status: tc.status, StatusMessage: "previous result", VerificationRequestedAt: tc.requested}
			if tc.status != app.CloudConnectionStatusPending {
				connection.LastVerifiedAt = &lastVerified
			}
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("load", func(tx *gorm.DB) {
				if dest, ok := tx.Statement.Dest.(*app.CloudConnection); ok {
					*dest = connection
				}
			}))
			updates, enqueues := 0, 0
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("save", func(tx *gorm.DB) {
				updates++
				update := tx.Statement.Dest.(app.CloudConnection)
				require.Equal(t, []string{"verification_requested_at"}, tx.Statement.Selects)
				require.NotNil(t, update.VerificationRequestedAt)
				require.Equal(t, now, *update.VerificationRequestedAt)
				require.Empty(t, update.Status)
				require.Empty(t, update.StatusMessage)
				require.Nil(t, update.LastVerifiedAt)
			}))
			svc := &service{db: db, now: func() time.Time { return now }, enqueueVerification: func(_ context.Context, c *app.CloudConnection) error {
				enqueues++
				require.Equal(t, connection, *c)
				return tc.enqueueError
			}}
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/cloud-connections/cc_acme/verify", nil)
			ctx.Params = gin.Params{{Key: "connection_id", Value: connection.ID}}
			cctx.SetOrgGinContext(ctx, &app.Org{ID: connection.OrgID})
			svc.Verify(ctx)
			wantEnqueues := 1
			if tc.deduplicated {
				wantEnqueues = 0
			}
			require.Equal(t, wantEnqueues, enqueues)
			if tc.enqueueError != nil {
				require.ErrorIs(t, ctx.Errors.Last().Err, tc.enqueueError)
				require.Zero(t, updates)
				return
			}
			require.Empty(t, ctx.Errors)
			require.Equal(t, http.StatusAccepted, w.Code)
			require.Equal(t, wantEnqueues, updates)
			var response ConnectionResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			require.Equal(t, connection.Status, response.Status)
			require.Equal(t, connection.StatusMessage, response.StatusMessage)
			require.Equal(t, connection.LastVerifiedAt, response.LastVerifiedAt)
			require.NotNil(t, response.VerificationRequestedAt)
			if tc.deduplicated {
				require.Equal(t, tc.requested, response.VerificationRequestedAt)
			} else {
				require.Equal(t, now, *response.VerificationRequestedAt)
			}
			require.True(t, response.VerificationInProgress)
		})
	}
}

type dryRunTransaction struct{ gorm.ConnPool }

func (tx dryRunTransaction) BeginTx(context.Context, *sql.TxOptions) (gorm.ConnPool, error) {
	return tx, nil
}

func (dryRunTransaction) Commit() error   { return nil }
func (dryRunTransaction) Rollback() error { return nil }

func TestDeleteConflict(t *testing.T) {
	for name, tc := range map[string]struct {
		installs int64
		message  string
	}{
		"one install":       {1, "This connection is used by 1 install. Delete that install first."},
		"multiple installs": {3, "This connection is used by 3 installs. Delete those installs first."},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
			require.NoError(t, err)
			db.Statement.ConnPool = dryRunTransaction{db.ConnPool}
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("load", func(tx *gorm.DB) {
				switch dest := tx.Statement.Dest.(type) {
				case *app.CloudConnection:
					*dest = app.CloudConnection{ID: "cc_acme", OrgID: "org_acme"}
				case *int64:
					*dest = tc.installs
					tx.RowsAffected = 1
				}
			}))
			require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("no-delete", func(tx *gorm.DB) {
				t.Error("referenced connection must not be deleted")
			}))
			svc := &service{db: db}
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodDelete, "/v1/cloud-connections/cc_acme", nil)
			ctx.Params = gin.Params{{Key: "connection_id", Value: "cc_acme"}}
			cctx.SetOrgGinContext(ctx, &app.Org{ID: "org_acme"})
			svc.Delete(ctx)
			require.Len(t, ctx.Errors, 1)
			var conflict stderr.ErrConflict
			require.ErrorAs(t, ctx.Errors.Last().Err, &conflict)
			require.Equal(t, tc.message, conflict.Description)
		})
	}
}
