package generics

import (
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type jsonTestRecord struct {
	ID     string
	Status []byte
}

func (jsonTestRecord) TableName() string {
	return "json_test_records"
}

type jsonSoftDeleteRecord struct {
	ID        string
	Status    []byte
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt
}

func (jsonSoftDeleteRecord) TableName() string {
	return "json_soft_delete_records"
}

func jsonTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true},
	)
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	return db
}

func TestJSONBStatusScopes(t *testing.T) {
	tests := map[string]struct {
		scope func(*gorm.DB) *gorm.DB
		want  string
		vars  []any
	}{
		"equal": {
			scope: WhereJSONBStatus("status", "active"),
			want:  "status->>'status' = $1",
			vars:  []any{"active"},
		},
		"not equal": {
			scope: WhereJSONBStatusNot("status", "error"),
			want:  "status->>'status' != $1",
			vars:  []any{"error"},
		},
		"in": {
			scope: WhereJSONBStatusIn("records.status", "active", "offline"),
			want:  "records.status->>'status' IN ($1,$2)",
			vars:  []any{"active", "offline"},
		},
		"not in": {
			scope: WhereJSONBStatusNotIn("status", "success", "error"),
			want:  "status->>'status' NOT IN ($1,$2)",
			vars:  []any{"success", "error"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tx := jsonTestDB(t).
				Model(&jsonTestRecord{}).
				Scopes(tt.scope).
				Find(&[]jsonTestRecord{})

			if got := tx.Statement.SQL.String(); !strings.Contains(got, tt.want) {
				t.Fatalf("SQL %q does not contain %q", got, tt.want)
			}
			if len(tx.Statement.Vars) != len(tt.vars) {
				t.Fatalf("got %d variables, want %d", len(tx.Statement.Vars), len(tt.vars))
			}
			for i := range tt.vars {
				if tx.Statement.Vars[i] != tt.vars[i] {
					t.Errorf("variable %d = %#v, want %#v", i, tx.Statement.Vars[i], tt.vars[i])
				}
			}
		})
	}
}

func TestSetJSONBMetadataKey(t *testing.T) {
	tx := SetJSONBMetadataKey(
		jsonTestDB(t).Model(&jsonTestRecord{}).Where(jsonTestRecord{ID: "record-id"}),
		"status",
		"shutdown_requested",
		true,
	)

	sql := tx.Statement.SQL.String()
	if !strings.Contains(sql, "jsonb_set(jsonb_set(COALESCE(status, '{}'::jsonb)") {
		t.Fatalf("SQL %q does not initialize the JSONB value and metadata object: %v", sql, tx.Error)
	}
	if strings.Contains(sql, "shutdown_requested") {
		t.Fatalf("SQL %q interpolates the metadata key", sql)
	}
	if len(tx.Statement.Vars) != 3 {
		t.Fatalf("got %d variables, want 3", len(tx.Statement.Vars))
	}
	path, ok := tx.Statement.Vars[0].(driver.Valuer)
	if !ok {
		t.Fatalf("path variable %#v does not implement driver.Valuer", tx.Statement.Vars[0])
	}
	pathValue, err := path.Value()
	if err != nil {
		t.Fatalf("encode path variable: %v", err)
	}
	pathString, ok := pathValue.(string)
	if !ok || !strings.Contains(pathString, "metadata") || !strings.Contains(pathString, "shutdown_requested") {
		t.Errorf("path variable = %#v, want metadata and shutdown_requested", pathValue)
	}
	if tx.Statement.Vars[1] != "true" {
		t.Errorf("value variable = %#v, want %q", tx.Statement.Vars[1], "true")
	}
	if tx.Statement.Vars[2] != "record-id" {
		t.Errorf("where variable = %#v, want %q", tx.Statement.Vars[2], "record-id")
	}
}

// On a soft-deleted model the deleted_at clause is the only condition gorm sees,
// so a bulk update needs either a real predicate or AllowGlobalUpdate. Without one
// it fails instead of updating, which is silent unless the error is checked.
func TestSetJSONBMetadataKeySoftDeleteGuard(t *testing.T) {
	t.Run("rejects update with no predicate", func(t *testing.T) {
		tx := SetJSONBMetadataKey(
			jsonTestDB(t).Model(&jsonSoftDeleteRecord{}),
			"status",
			"restart_hint",
			"now",
		)

		if !errors.Is(tx.Error, gorm.ErrMissingWhereClause) {
			t.Fatalf("error = %v, want %v", tx.Error, gorm.ErrMissingWhereClause)
		}
	})

	t.Run("allows deliberate global update", func(t *testing.T) {
		tx := SetJSONBMetadataKey(
			jsonTestDB(t).
				Session(&gorm.Session{AllowGlobalUpdate: true}).
				Model(&jsonSoftDeleteRecord{}),
			"status",
			"restart_hint",
			"now",
		)

		if tx.Error != nil {
			t.Fatalf("unexpected error: %v", tx.Error)
		}
		if got := tx.Statement.SQL.String(); !strings.Contains(got, `"deleted_at" = $`) {
			t.Errorf("SQL %q does not scope to live rows", got)
		}
	})

	t.Run("does not bump updated_at", func(t *testing.T) {
		tx := SetJSONBMetadataKey(
			jsonTestDB(t).
				Model(&jsonSoftDeleteRecord{}).
				Where(jsonSoftDeleteRecord{ID: "record-id"}),
			"status",
			"restart_hint",
			"now",
		)

		if tx.Error != nil {
			t.Fatalf("unexpected error: %v", tx.Error)
		}
		if got := tx.Statement.SQL.String(); strings.Contains(got, "updated_at") {
			t.Errorf("SQL %q writes updated_at", got)
		}
	})
}

func TestDeleteJSONBMetadataKey(t *testing.T) {
	tx := DeleteJSONBMetadataKey(
		jsonTestDB(t).Model(&jsonTestRecord{}).Where(jsonTestRecord{ID: "record-id"}),
		"status",
		"restart_hint",
	)

	if got := tx.Statement.SQL.String(); !strings.Contains(got, "COALESCE(status->'metadata', '{}'::jsonb) - $1") {
		t.Fatalf("SQL %q does not remove the metadata key: %v", got, tx.Error)
	}
}
