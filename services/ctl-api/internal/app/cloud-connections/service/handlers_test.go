package service

import (
	"context"
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
	for name, tc := range map[string]struct {
		status       app.CloudConnectionStatus
		enqueueError error
	}{
		"pending":       {status: app.CloudConnectionStatusPending},
		"verified":      {status: app.CloudConnectionStatusVerified},
		"error":         {status: app.CloudConnectionStatusError},
		"enqueue fails": {status: app.CloudConnectionStatusVerified, enqueueError: errors.New("queue full")},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost dbname=unused"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			lastVerified := time.Now().UTC().Add(-time.Hour)
			connection := app.CloudConnection{ID: "cc_acme", OrgID: "org_acme", Status: tc.status, StatusMessage: "previous result"}
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
				require.WithinDuration(t, time.Now(), *update.VerificationRequestedAt, time.Second)
				require.Empty(t, update.Status)
				require.Empty(t, update.StatusMessage)
				require.Nil(t, update.LastVerifiedAt)
			}))
			svc := &service{db: db, enqueueVerification: func(_ context.Context, c *app.CloudConnection) error {
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
			require.Equal(t, 1, enqueues)
			if tc.enqueueError != nil {
				require.ErrorIs(t, ctx.Errors.Last().Err, tc.enqueueError)
				require.Zero(t, updates)
				return
			}
			require.Empty(t, ctx.Errors)
			require.Equal(t, http.StatusAccepted, w.Code)
			require.Equal(t, 1, updates)
			var response ConnectionResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			require.Equal(t, connection.Status, response.Status)
			require.Equal(t, connection.StatusMessage, response.StatusMessage)
			require.Equal(t, connection.LastVerifiedAt, response.LastVerifiedAt)
			require.NotNil(t, response.VerificationRequestedAt)
			require.True(t, response.VerificationInProgress)
		})
	}
}
