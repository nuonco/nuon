package configdiff

import (
	"testing"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestComponentDiffEntryBuildChangedSameChecksum(t *testing.T) {
	oldConn := &app.ComponentConfigConnection{
		ComponentID:   "comp-1",
		ComponentName: "api",
		Type:          app.ComponentTypeDockerBuild,
		Checksum:      "abc123",
		LatestBuildID: generics.NewNullString("bld-old"),
	}
	newConn := &app.ComponentConfigConnection{
		ComponentID:   "comp-1",
		ComponentName: "api",
		Type:          app.ComponentTypeDockerBuild,
		Checksum:      "abc123",
		LatestBuildID: generics.NewNullString("bld-new"),
	}

	entry := componentDiffEntry(oldConn, newConn)
	if !entry.BuildChanged {
		t.Fatal("expected build_changed when checksums match but build IDs differ")
	}
	if entry.OldBuildID != "bld-old" || entry.NewBuildID != "bld-new" {
		t.Fatalf("unexpected build IDs: old=%q new=%q", entry.OldBuildID, entry.NewBuildID)
	}
}

func TestComponentDiffEntryUnchangedChecksumAndBuild(t *testing.T) {
	oldConn := &app.ComponentConfigConnection{
		ComponentID:   "comp-1",
		Checksum:      "abc123",
		LatestBuildID: generics.NewNullString("bld-same"),
	}
	newConn := &app.ComponentConfigConnection{
		ComponentID:   "comp-1",
		Checksum:      "abc123",
		LatestBuildID: generics.NewNullString("bld-same"),
	}

	entry := componentDiffEntry(oldConn, newConn)
	if entry.BuildChanged {
		t.Fatal("expected no build_changed when checksum and build IDs match")
	}
}

func TestChecksumsEqualRequiresBothNonEmpty(t *testing.T) {
	oldConn := &app.ComponentConfigConnection{Checksum: "abc"}
	newConn := &app.ComponentConfigConnection{Checksum: ""}
	if checksumsEqual(oldConn, newConn) {
		t.Fatal("empty checksum should not compare equal")
	}
}
