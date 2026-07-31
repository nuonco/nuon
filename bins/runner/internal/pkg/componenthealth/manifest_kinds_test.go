package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const nodePoolManifest = `
---
# Source: chart/templates/nodepool.yaml
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: general
---
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: general
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: marker
`

func TestGvksFromManifest(t *testing.T) {
	assert.Equal(t, []schema.GroupVersionKind{
		{Group: "karpenter.sh", Version: "v1", Kind: "NodePool"},
		{Group: "karpenter.k8s.aws", Version: "v1", Kind: "EC2NodeClass"},
		{Group: "", Version: "v1", Kind: "ConfigMap"},
	}, gvksFromManifest(nodePoolManifest))

	assert.Empty(t, gvksFromManifest(""))
	assert.Empty(t, gvksFromManifest("not a manifest"))
	// A document missing either field is skipped rather than half-guessed.
	assert.Empty(t, gvksFromManifest("---\nkind: NodePool\n"))
	assert.Empty(t, gvksFromManifest("---\napiVersion: v1\n"))
}

func TestGvksFromManifestDeduplicates(t *testing.T) {
	m := "---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: a\n---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: b\n"
	assert.Len(t, gvksFromManifest(m), 1, "the same kind twice is one kind to watch")
}

func TestManifestKindsProviderSetAndDiscover(t *testing.T) {
	p := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop()})

	p.Set("cmp-a", nodePoolManifest)
	assert.Len(t, p.DiscoveredGVKs(), 3)

	// A chart that no longer renders anything stops contributing kinds, so a
	// removed CR does not keep costing a list call every cycle.
	p.Set("cmp-a", "")
	assert.Empty(t, p.DiscoveredGVKs())

	p.Set("", nodePoolManifest)
	assert.Empty(t, p.DiscoveredGVKs(), "an empty component id records nothing")
}

func TestDecodeComponentKind(t *testing.T) {
	id, gvk, ok := decodeComponentKind("cmp1|karpenter.sh/v1/NodePool")
	assert.True(t, ok)
	assert.Equal(t, "cmp1", id)
	assert.Equal(t, schema.GroupVersionKind{Group: "karpenter.sh", Version: "v1", Kind: "NodePool"}, gvk)

	// Core kinds have an empty group.
	_, gvk, ok = decodeComponentKind("cmp1|v1/ConfigMap")
	assert.True(t, ok)
	assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, gvk)

	for _, bad := range []string{"", "no-pipe", "cmp1|", "|v1/ConfigMap", "cmp1|NodePool", "cmp1|v1/"} {
		_, _, ok := decodeComponentKind(bad)
		assert.False(t, ok, bad)
	}
}

// A restart must not narrow the watch set: kinds persisted by earlier deploys
// come back without waiting for every component to redeploy.
func TestManifestKindsRoundTripsThroughPersistence(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}

	first := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	first.Set("cmp-a", nodePoolManifest)
	assert.Len(t, first.DiscoveredGVKs(), 3)
	assert.Len(t, store.ComponentKinds(), 3, "kinds should have been handed to the store")

	// A fresh process, same persisted context.
	restarted := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	assert.Empty(t, restarted.DiscoveredGVKs(), "nothing until it loads")
	restarted.Load()
	assert.ElementsMatch(t, first.DiscoveredGVKs(), restarted.DiscoveredGVKs())
}
