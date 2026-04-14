package service

import (
	"strings"
	"testing"
	"time"
)

func TestBuildBaseWhereClause(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		filter      analyticsFilter
		wantClauses []string
		wantParams  int
	}{
		{
			name: "no optional filters",
			filter: analyticsFilter{
				OrgID: "org1", AppID: "app1",
				Start: start, End: end,
			},
			wantClauses: []string{"org_id = ?", "app_id = ?", "evaluated_at BETWEEN"},
			wantParams:  4,
		},
		{
			name: "with install_id",
			filter: analyticsFilter{
				OrgID: "org1", AppID: "app1",
				Start: start, End: end,
				InstallID: "inst1",
			},
			wantClauses: []string{"install_id = ?"},
			wantParams:  5,
		},
		{
			name: "with both filters",
			filter: analyticsFilter{
				OrgID: "org1", AppID: "app1",
				Start: start, End: end,
				InstallID: "inst1", PolicyID: "pol1",
			},
			wantClauses: []string{"install_id = ?", "policy_id = ?"},
			wantParams:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := buildBaseWhereClause(tt.filter)

			for _, clause := range tt.wantClauses {
				if !strings.Contains(query, clause) {
					t.Errorf("query missing clause %q: %s", clause, query)
				}
			}
			if len(params) != tt.wantParams {
				t.Errorf("params count = %d, want %d", len(params), tt.wantParams)
			}
		})
	}
}

func TestBuildTimeseriesSelectClauses(t *testing.T) {
	interval := timeInterval{"day", "toStartOfDay(%s)"}

	t.Run("no dimensions", func(t *testing.T) {
		sel, grp, ord := buildTimeseriesSelectClauses(interval, nil)

		if !strings.Contains(sel, "toStartOfDay(evaluated_at) AS bucket") {
			t.Errorf("select missing bucket expr: %s", sel)
		}
		if !strings.Contains(sel, "countIf(outcome = 'deny')") {
			t.Errorf("select missing countIf: %s", sel)
		}
		if grp != "bucket" {
			t.Errorf("groupCols = %q, want %q", grp, "bucket")
		}
		if ord != "bucket" {
			t.Errorf("orderCols = %q, want %q", ord, "bucket")
		}
		if strings.Contains(sel, "policy_id") {
			t.Error("should not contain dimension columns when no dims")
		}
	})

	t.Run("single dimension", func(t *testing.T) {
		sel, grp, ord := buildTimeseriesSelectClauses(interval, []string{"policy_id"})

		if !strings.Contains(sel, "policy_id") {
			t.Errorf("select missing policy_id column: %s", sel)
		}
		if grp != "bucket, policy_id" {
			t.Errorf("groupCols = %q, want %q", grp, "bucket, policy_id")
		}
		if ord != "bucket, policy_id" {
			t.Errorf("orderCols = %q, want %q", ord, "bucket, policy_id")
		}
	})

	t.Run("multiple dimensions", func(t *testing.T) {
		sel, grp, ord := buildTimeseriesSelectClauses(interval, []string{"policy_id", "install_id"})

		if !strings.Contains(sel, "policy_id") || !strings.Contains(sel, "install_id") {
			t.Errorf("select missing dimension columns: %s", sel)
		}
		if grp != "bucket, policy_id, install_id" {
			t.Errorf("groupCols = %q, want %q", grp, "bucket, policy_id, install_id")
		}
		if ord != "bucket, policy_id, install_id" {
			t.Errorf("orderCols = %q, want %q", ord, "bucket, policy_id, install_id")
		}
	})
}

func TestIsValidGroupBy(t *testing.T) {
	valid := []string{"policy_id", "install_id", "component_id", "owner_type"}
	for _, v := range valid {
		if _, ok := validGroupByDimensions[v]; !ok {
			t.Errorf("validGroupByDimensions missing %q", v)
		}
	}

	invalid := []string{"org_id", "SELECT 1", "policy_id; DROP TABLE"}
	for _, v := range invalid {
		if _, ok := validGroupByDimensions[v]; ok {
			t.Errorf("validGroupByDimensions should not contain %q", v)
		}
	}
}
