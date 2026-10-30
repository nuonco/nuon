package v2

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestFilterComponentsByDiffEligibleOrderProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		n := hegel.Draw(t, hegel.Integers(3, 9))
		cfg := &app.AppConfig{}
		diff := &app.InstallConfigDiff{}
		members, eligible := map[string]bool{}, map[string]bool{}
		var order []string
		for i := range n {
			id := fmt.Sprintf("component-%d", i)
			present := hegel.Draw(t, hegel.Booleans())
			applied := hegel.Draw(t, hegel.Booleans())
			appliedRevision := hegel.Draw(t, hegel.Integers(0, 2))
			targetRevision := hegel.Draw(t, hegel.Integers(0, 2))
			impact := hegel.Draw(t, hegel.Booleans())
			buildOnly := hegel.Draw(t, hegel.Booleans())
			toggle := hegel.Draw(t, hegel.SampledFrom([]int{0, 1, 2}))
			conn := app.ComponentConfigConnection{ComponentID: id, DefaultEnabled: generics.ToPtr(hegel.Draw(t, hegel.Booleans()))}
			if toggle != 0 {
				conn.Toggleable = generics.ToPtr(toggle == 1)
			}
			members[id] = present
			if present {
				cfg.ComponentIDs = append(cfg.ComponentIDs, id)
				cfg.ComponentConfigConnections = append(cfg.ComponentConfigConnections, conn)
			}
			entry := app.ComponentDiffEntry{ComponentID: id, BuildChanged: buildOnly}
			switch {
			case !present:
				diff.Removed = append(diff.Removed, entry)
			case !applied:
				diff.Added = append(diff.Added, entry)
				eligible[id] = true
			case appliedRevision != targetRevision || impact:
				diff.Changed = append(diff.Changed, entry)
				eligible[id] = true
			default:
				diff.Unchanged = append(diff.Unchanged, entry)
			}
			if hegel.Draw(t, hegel.Booleans()) {
				order = append(order, id)
			}
		}
		cfg.ComponentIDs = append(cfg.ComponentIDs, "disabled-added", "build-only")
		cfg.ComponentConfigConnections = append(cfg.ComponentConfigConnections, app.ComponentConfigConnection{
			ComponentID: "disabled-added", Toggleable: generics.ToPtr(true), DefaultEnabled: generics.ToPtr(false),
		})
		diff.Added = append(diff.Added, app.ComponentDiffEntry{ComponentID: "disabled-added"})
		diff.Changed = append(diff.Changed, app.ComponentDiffEntry{ComponentID: "not-in-target"})
		diff.Unchanged = append(diff.Unchanged, app.ComponentDiffEntry{ComponentID: "build-only", BuildChanged: true})
		members["disabled-added"], members["build-only"], eligible["disabled-added"] = true, true, true
		order = append(order, "disabled-added", "build-only", "not-in-target")
		for i := len(order) - 1; i > 0; i-- {
			j := hegel.Draw(t, hegel.Integers(0, i))
			order[i], order[j] = order[j], order[i]
		}
		orderBefore := slices.Clone(order)
		memberBefore := slices.Clone(cfg.ComponentIDs)
		var want []string
		for _, id := range order {
			if members[id] && eligible[id] {
				want = append(want, id)
			}
		}
		got := filterComponentsByDiff(order, cfg, diff)
		require.Equal(t, want, got)
		require.Equal(t, got, filterComponentsByDiff(order, cfg, diff))
		require.Equal(t, got, filterComponentsByDiff(got, cfg, diff))
		require.NotContains(t, got, "build-only")
		require.Contains(t, got, "disabled-added")
		require.NotContains(t, got, "not-in-target")
		var full []string
		for _, id := range order {
			if members[id] {
				full = append(full, id)
			}
		}
		require.Equal(t, full, filterComponentsByDiff(order, cfg, nil))
		for i := range cfg.ComponentConfigConnections {
			cfg.ComponentConfigConnections[i].Toggleable = generics.ToPtr(true)
			cfg.ComponentConfigConnections[i].DefaultEnabled = generics.ToPtr(false)
		}
		require.Equal(t, want, filterComponentsByDiff(order, cfg, diff))
		require.Equal(t, orderBefore, order)
		require.Equal(t, memberBefore, cfg.ComponentIDs)
	}, hegel.WithTestCases(250))
}
