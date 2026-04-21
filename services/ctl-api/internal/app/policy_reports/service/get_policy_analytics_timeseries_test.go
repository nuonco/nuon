package service

import (
	"testing"
	"time"
)

func TestBuildTimeseriesBuckets_NoDimensions(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, Evaluations: 10, Denies: 2, Warns: 3, Passes: 5},
		{Bucket: t2, Evaluations: 8, Denies: 0, Warns: 1, Passes: 7},
	}

	buckets := buildTimeseriesBuckets(rows, nil)

	if len(buckets) != 2 {
		t.Fatalf("got %d buckets, want 2", len(buckets))
	}
	if buckets[0].Denies != 2 {
		t.Errorf("bucket[0].Denies = %d, want 2", buckets[0].Denies)
	}
	if buckets[1].Passes != 7 {
		t.Errorf("bucket[1].Passes = %d, want 7", buckets[1].Passes)
	}
	if buckets[0].Series != nil {
		t.Error("Series should be nil when no dimensions")
	}
}

func TestBuildTimeseriesBuckets_SingleDimension(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, PolicyID: "pol_a", Evaluations: 5, Denies: 1, Warns: 0, Passes: 4},
		{Bucket: t1, PolicyID: "pol_b", Evaluations: 3, Denies: 0, Warns: 1, Passes: 2},
	}

	buckets := buildTimeseriesBuckets(rows, []string{"policy_id"})

	if len(buckets) != 1 {
		t.Fatalf("got %d buckets, want 1", len(buckets))
	}

	b := buckets[0]
	if b.Evaluations != 8 {
		t.Errorf("aggregated Evaluations = %d, want 8", b.Evaluations)
	}
	if len(b.Series) != 2 {
		t.Fatalf("got %d series, want 2", len(b.Series))
	}
	if b.Series[0].Labels["policy_id"] != "pol_a" {
		t.Errorf("series[0] policy_id = %q, want %q", b.Series[0].Labels["policy_id"], "pol_a")
	}
	if b.Series[0].Denies != 1 {
		t.Errorf("series[0] Denies = %d, want 1", b.Series[0].Denies)
	}
}

func TestBuildTimeseriesBuckets_MultiDimension(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, PolicyID: "pol_a", InstallID: "inst_1", Evaluations: 3, Denies: 1},
		{Bucket: t1, PolicyID: "pol_a", InstallID: "inst_2", Evaluations: 2, Denies: 0},
		{Bucket: t1, PolicyID: "pol_b", InstallID: "inst_1", Evaluations: 4, Denies: 2},
	}

	dims := []string{"policy_id", "install_id"}
	buckets := buildTimeseriesBuckets(rows, dims)

	if len(buckets) != 1 {
		t.Fatalf("got %d buckets, want 1", len(buckets))
	}

	b := buckets[0]
	if b.Evaluations != 9 {
		t.Errorf("aggregated Evaluations = %d, want 9", b.Evaluations)
	}
	if b.Denies != 3 {
		t.Errorf("aggregated Denies = %d, want 3", b.Denies)
	}
	if len(b.Series) != 3 {
		t.Fatalf("got %d series, want 3", len(b.Series))
	}

	// Verify multi-label structure
	s := b.Series[0]
	if len(s.Labels) != 2 {
		t.Errorf("series[0] has %d labels, want 2", len(s.Labels))
	}
	if s.Labels["policy_id"] != "pol_a" || s.Labels["install_id"] != "inst_1" {
		t.Errorf("series[0] labels = %v, want policy_id=pol_a, install_id=inst_1", s.Labels)
	}
}

func TestBuildTimeseriesBuckets_Empty(t *testing.T) {
	buckets := buildTimeseriesBuckets(nil, nil)
	if len(buckets) != 0 {
		t.Errorf("got %d buckets, want 0", len(buckets))
	}

	buckets = buildTimeseriesBuckets(nil, []string{"policy_id"})
	if len(buckets) != 0 {
		t.Errorf("got %d buckets (grouped), want 0", len(buckets))
	}
}

func TestBuildTimeseriesBuckets_OrderPreserved(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, PolicyID: "a", Evaluations: 1},
		{Bucket: t2, PolicyID: "a", Evaluations: 2},
		{Bucket: t3, PolicyID: "a", Evaluations: 3},
		{Bucket: t1, PolicyID: "b", Evaluations: 4},
	}

	buckets := buildTimeseriesBuckets(rows, []string{"policy_id"})

	if len(buckets) != 3 {
		t.Fatalf("got %d buckets, want 3", len(buckets))
	}
	if !buckets[0].Time.Equal(t1) || !buckets[1].Time.Equal(t2) || !buckets[2].Time.Equal(t3) {
		t.Error("bucket order not preserved")
	}
	if buckets[0].Evaluations != 5 {
		t.Errorf("bucket[0] aggregated Evaluations = %d, want 5 (1+4)", buckets[0].Evaluations)
	}
	if len(buckets[0].Series) != 2 {
		t.Errorf("bucket[0] series count = %d, want 2", len(buckets[0].Series))
	}
}

func TestParseGroupBy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"single", "policy_id", []string{"policy_id"}, false},
		{"multiple", "policy_id,install_id", []string{"policy_id", "install_id"}, false},
		{"with spaces", " policy_id , install_id ", []string{"policy_id", "install_id"}, false},
		{"deduplicates", "policy_id,policy_id", []string{"policy_id"}, false},
		{"all valid", "policy_id,install_id,component_id,owner_type", []string{"policy_id", "install_id", "component_id", "owner_type"}, false},
		{"invalid dimension", "policy_id,bad_field", nil, true},
		{"sql injection", "policy_id; DROP TABLE", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGroupBy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("got %v, want %v", got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestExtractLabels(t *testing.T) {
	row := timeseriesRow{
		PolicyID:    "pol_abc",
		InstallID:   "inst_1",
		ComponentID: "comp_x",
		OwnerType:   "install_deploys",
	}

	t.Run("single dimension", func(t *testing.T) {
		labels := extractLabels(row, []string{"policy_id"})
		if len(labels) != 1 || labels["policy_id"] != "pol_abc" {
			t.Errorf("labels = %v, want {policy_id: pol_abc}", labels)
		}
	})

	t.Run("multiple dimensions", func(t *testing.T) {
		labels := extractLabels(row, []string{"policy_id", "install_id", "owner_type"})
		if len(labels) != 3 {
			t.Errorf("got %d labels, want 3", len(labels))
		}
		if labels["install_id"] != "inst_1" {
			t.Errorf("install_id = %q, want %q", labels["install_id"], "inst_1")
		}
		if labels["owner_type"] != "install_deploys" {
			t.Errorf("owner_type = %q, want %q", labels["owner_type"], "install_deploys")
		}
	})
}
