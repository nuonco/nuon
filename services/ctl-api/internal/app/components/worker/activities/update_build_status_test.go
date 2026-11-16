package activities

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type dryRunConnPool struct{}

func (dryRunConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, nil
}

func (dryRunConnPool) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return sql.Result(nil), nil
}

func (dryRunConnPool) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (dryRunConnPool) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return nil
}

func (dryRunConnPool) BeginTx(context.Context, *sql.TxOptions) (gorm.ConnPool, error) {
	return dryRunConnPool{}, nil
}

func (dryRunConnPool) Commit() error { return nil }

func (dryRunConnPool) Rollback() error { return nil }

func TestUpdateBuildFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		previous    *compositeerrors.CompositeErrorData
		policy      bool
		cancelled   bool
		queryError  error
		updateError error
		missing     bool
		wantError   string
	}{
		"failure before a runner result exists": {},
		"existing structured failure": {
			previous: &compositeerrors.CompositeErrorData{Version: 1, Type: "oci.image-not-found", Message: "Image reference not found"},
		},
		"policy failure":  {policy: true},
		"cancelled build": {cancelled: true},
		"read failure":    {queryError: errors.New("database read unavailable"), wantError: "database read unavailable"},
		"write failure":   {updateError: errors.New("database write unavailable"), wantError: "database write unavailable"},
		"missing build":   {missing: true, wantError: "no build found"},
	} {
		t.Run(name, func(t *testing.T) {
			db, err := gorm.Open(postgres.New(postgres.Config{Conn: dryRunConnPool{}}), &gorm.Config{
				DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("load", func(tx *gorm.DB) {
				if tc.queryError != nil {
					tx.AddError(tc.queryError)
					return
				}
				build := tx.Statement.Dest.(*app.ComponentBuild)
				build.CompositeError = tc.previous
				if tc.policy {
					build.Status = app.ComponentBuildStatusPolicyFailed
					build.StatusDescription = "policy violations: image must be signed"
				}
				if tc.cancelled {
					build.Status = app.ComponentBuildStatus(app.StatusCancelled)
					build.StatusDescription = "build cancelled"
				}
			}))
			var stored app.ComponentBuild
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("save", func(tx *gorm.DB) {
				stored = tx.Statement.Dest.(app.ComponentBuild)
				tx.RowsAffected = 1
				if tc.missing {
					tx.RowsAffected = 0
				}
				if tc.updateError != nil {
					tx.AddError(tc.updateError)
				}
			}))
			err = (&Activities{db: db}).UpdateBuildStatus(context.Background(), UpdateBuildStatus{
				BuildID: "build-id", Status: app.ComponentBuildStatusError,
				StatusDescription: "Build failed: UNAUTHORIZED registry denied password=acme-secret",
				Error:             "unable to fetch image metadata: UNAUTHORIZED registry denied password=acme-secret client_assertion=acme-assertion",
			})
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, stored.CompositeError)
			encoded, err := json.Marshal(stored)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "acme-secret")
			require.NotContains(t, string(encoded), "acme-assertion")
			if tc.policy {
				require.Equal(t, app.ComponentBuildStatusPolicyFailed, stored.Status)
				require.Equal(t, "policy violations: image must be signed", stored.StatusDescription)
			} else if tc.cancelled {
				require.Equal(t, app.ComponentBuildStatus(app.StatusCancelled), stored.Status)
				require.Equal(t, "build cancelled", stored.StatusDescription)
			} else {
				require.Equal(t, app.ComponentBuildStatusError, stored.Status)
			}
			if tc.previous != nil {
				require.Equal(t, tc.previous, stored.CompositeError)
			} else {
				require.Equal(t, stored.StatusDescription, stored.CompositeError.Message)
				require.Contains(t, string(encoded), "UNAUTHORIZED registry denied")
				require.Contains(t, string(encoded), "unable to fetch image metadata")
			}
		})
	}
}
