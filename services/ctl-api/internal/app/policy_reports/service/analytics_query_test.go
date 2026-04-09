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

	t.Run("without group_by", func(t *testing.T) {
		sel, grp, ord := buildTimeseriesSelectClauses(interval, "")

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
	})

	t.Run("with group_by policy_id", func(t *testing.T) {
		sel, grp, ord := buildTimeseriesSelectClauses(interval, "policy_id")

		if !strings.Contains(sel, "policy_id AS group_key") {
			t.Errorf("select missing group_key: %s", sel)
		}
		if grp != "bucket, group_key" {
			t.Errorf("groupCols = %q, want %q", grp, "bucket, group_key")
		}
		if ord != "bucket, group_key" {
			t.Errorf("orderCols = %q, want %q", ord, "bucket, group_key")
		}
	})
}

func TestIsValidGroupBy(t *testing.T) {
	valid := []string{"policy_id", "install_id", "component_id"}
	for _, v := range valid {
		if !isValidGroupBy(v) {
			t.Errorf("isValidGroupBy(%q) = false, want true", v)
		}
	}

	invalid := []string{"", "org_id", "SELECT 1", "policy_id; DROP TABLE"}
	for _, v := range invalid {
		if isValidGroupBy(v) {
			t.Errorf("isValidGroupBy(%q) = true, want false", v)
		}
	}
}
