package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
)

func TestAddHistoricalActionDefinitions(t *testing.T) {
	definition := &operation.BundleActionDefinition{Steps: []operation.BundleActionStep{{Name: "restart", Command: "kubectl rollout restart"}}}
	history := []operation.BundleInfo{{
		ArchiveDigest: "sha256:archive",
		Contents: []operation.BundleContent{
			{Kind: operation.BundleContentKindAction, Name: "restart"},
			{Kind: operation.BundleContentKindComponent, Name: "api"},
		},
	}}

	addHistoricalActionDefinitions(history, bundleActionDefinitions{
		"sha256:archive": {"restart": definition},
	})

	require.Equal(t, definition, history[0].Contents[0].ActionDefinition)
	require.Nil(t, history[0].Contents[1].ActionDefinition)
}
