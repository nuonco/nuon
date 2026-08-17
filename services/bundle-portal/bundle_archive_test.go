package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

func TestAddHistoricalActionDefinitions(t *testing.T) {
	definition := &day2.BundleActionDefinition{Steps: []day2.BundleActionStep{{Name: "restart", Command: "kubectl rollout restart"}}}
	history := []day2.BundleInfo{{
		ArchiveDigest: "sha256:archive",
		Contents: []day2.BundleContent{
			{Kind: day2.BundleContentKindAction, Name: "restart"},
			{Kind: day2.BundleContentKindComponent, Name: "api"},
		},
	}}

	addHistoricalActionDefinitions(history, bundleActionDefinitions{
		"sha256:archive": {"restart": definition},
	})

	require.Equal(t, definition, history[0].Contents[0].ActionDefinition)
	require.Nil(t, history[0].Contents[1].ActionDefinition)
}
