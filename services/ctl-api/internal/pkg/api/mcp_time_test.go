package api

import (
	"testing"
	"time"
)

func TestMCPTimeUTC(t *testing.T) {
	got := MCPTime(time.Date(2026, 9, 4, 4, 23, 0, 0, time.UTC))
	if got != "2026-09-04T04:23:00Z" {
		t.Fatalf("got %q", got)
	}

	pst := time.FixedZone("PST", -8*60*60)
	got = MCPTime(time.Date(2026, 9, 3, 21, 23, 0, 0, pst))
	if got != "2026-09-04T05:23:00Z" {
		t.Fatalf("localized input got %q", got)
	}
}
