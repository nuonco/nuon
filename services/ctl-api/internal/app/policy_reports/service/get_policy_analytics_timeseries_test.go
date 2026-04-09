package service

import (
	"testing"
	"time"
)

func TestBuildTimeseriesBuckets_NoGroups(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, Evaluations: 10, Denies: 2, Warns: 3, Passes: 5},
		{Bucket: t2, Evaluations: 8, Denies: 0, Warns: 1, Passes: 7},
	}

	buckets := buildTimeseriesBuckets(rows, false)

	if len(buckets) != 2 {
		t.Fatalf("got %d buckets, want 2", len(buckets))
	}
	if buckets[0].Time != t1 {
		t.Errorf("bucket[0].Time = %v, want %v", buckets[0].Time, t1)
	}
	if buckets[0].Denies != 2 {
		t.Errorf("bucket[0].Denies = %d, want 2", buckets[0].Denies)
	}
	if buckets[1].Passes != 7 {
		t.Errorf("bucket[1].Passes = %d, want 7", buckets[1].Passes)
	}
	if buckets[0].Groups != nil {
		t.Error("bucket[0].Groups should be nil when hasGroups=false")
	}
}

func TestBuildTimeseriesBuckets_WithGroups(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, GroupKey: "pol_a", Evaluations: 5, Denies: 1, Warns: 0, Passes: 4},
		{Bucket: t1, GroupKey: "pol_b", Evaluations: 3, Denies: 0, Warns: 1, Passes: 2},
	}

	buckets := buildTimeseriesBuckets(rows, true)

	if len(buckets) != 1 {
		t.Fatalf("got %d buckets, want 1", len(buckets))
	}

	b := buckets[0]
	if b.Evaluations != 8 {
		t.Errorf("aggregated Evaluations = %d, want 8", b.Evaluations)
	}
	if b.Denies != 1 {
		t.Errorf("aggregated Denies = %d, want 1", b.Denies)
	}
	if b.Warns != 1 {
		t.Errorf("aggregated Warns = %d, want 1", b.Warns)
	}
	if b.Passes != 6 {
		t.Errorf("aggregated Passes = %d, want 6", b.Passes)
	}

	if len(b.Groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(b.Groups))
	}
	if b.Groups["pol_a"].Denies != 1 {
		t.Errorf("group pol_a Denies = %d, want 1", b.Groups["pol_a"].Denies)
	}
	if b.Groups["pol_b"].Warns != 1 {
		t.Errorf("group pol_b Warns = %d, want 1", b.Groups["pol_b"].Warns)
	}
}

func TestBuildTimeseriesBuckets_Empty(t *testing.T) {
	buckets := buildTimeseriesBuckets(nil, false)
	if len(buckets) != 0 {
		t.Errorf("got %d buckets, want 0", len(buckets))
	}

	buckets = buildTimeseriesBuckets(nil, true)
	if len(buckets) != 0 {
		t.Errorf("got %d buckets (grouped), want 0", len(buckets))
	}
}

func TestBuildTimeseriesBuckets_OrderPreserved(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	rows := []timeseriesRow{
		{Bucket: t1, GroupKey: "a", Evaluations: 1},
		{Bucket: t2, GroupKey: "a", Evaluations: 2},
		{Bucket: t3, GroupKey: "a", Evaluations: 3},
		{Bucket: t1, GroupKey: "b", Evaluations: 4},
	}

	buckets := buildTimeseriesBuckets(rows, true)

	if len(buckets) != 3 {
		t.Fatalf("got %d buckets, want 3", len(buckets))
	}
	if !buckets[0].Time.Equal(t1) || !buckets[1].Time.Equal(t2) || !buckets[2].Time.Equal(t3) {
		t.Error("bucket order not preserved")
	}
	if buckets[0].Evaluations != 5 {
		t.Errorf("bucket[0] aggregated Evaluations = %d, want 5 (1+4)", buckets[0].Evaluations)
	}
}
