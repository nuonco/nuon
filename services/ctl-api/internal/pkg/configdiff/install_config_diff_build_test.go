package configdiff

import (
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
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

func TestComponentBuildSelectionIndependentOfCategoryProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		impacted := hegel.Draw(t, hegel.Booleans())
		for _, oldChecksum := range []string{"", "config-a", "config-b"} {
			for _, newChecksum := range []string{"", "config-a", "config-b"} {
				for _, oldBuildID := range []string{"", "build-a", "build-b"} {
					for _, newBuildID := range []string{"", "build-a", "build-b"} {
						oldConn := app.ComponentConfigConnection{
							ComponentID: "cmp-api", ComponentName: "api", Type: app.ComponentTypeDockerBuild,
							Checksum: oldChecksum, LatestBuildID: generics.NewNullString(oldBuildID),
						}
						newConn := oldConn
						newConn.Checksum, newConn.LatestBuildID = newChecksum, generics.NewNullString(newBuildID)
						for _, oldPresent := range []bool{false, true} {
							loaded := map[string]*app.AppConfig{}
							if oldPresent {
								loaded["applied"] = &app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{oldConn}}
							}
							diff := &app.InstallConfigDiff{}
							err := fillComponentSlices(diff, CompositeBaselines{Fallback: "applied"}, loaded,
								&app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{newConn}},
								func(string) (*pkgdiff.Diff, error) {
									return &pkgdiff.Diff{ResourceID: "component.api", Impacted: impacted,
										ImpactReasons: []pkgdiff.ImpactReason{{From: "component.dependency"}}}, nil
								})
							require.NoError(t, err)
							var entries []app.ComponentDiffEntry
							switch {
							case !oldPresent:
								entries = diff.Added
							case oldChecksum != "" && oldChecksum == newChecksum && !impacted:
								entries = diff.Unchanged
							default:
								entries = diff.Changed
							}
							require.Len(t, entries, 1, "old_present=%t, checksums=%q/%q, ids=%q/%q", oldPresent, oldChecksum, newChecksum, oldBuildID, newBuildID)
							require.Equal(t, 1, len(diff.Added)+len(diff.Changed)+len(diff.Unchanged)+len(diff.Removed))
							wantOldID := ""
							if oldPresent {
								wantOldID = oldBuildID
							}
							require.Equal(t, wantOldID, entries[0].OldBuildID)
							require.Equal(t, newBuildID, entries[0].NewBuildID)
							require.Equal(t, wantOldID != newBuildID, entries[0].BuildChanged,
								"old_present=%t, checksums=%q/%q, ids=%q/%q", oldPresent, oldChecksum, newChecksum, oldBuildID, newBuildID)
							newConn.LatestBuildID.Valid = false
							require.Empty(t, componentDiffEntry(&oldConn, &newConn).NewBuildID)
							newConn.LatestBuildID.Valid = true
						}
					}
				}
			}
		}
	})
}
