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

	_, gvk, ok = decodeComponentKind("cmp1|v1/ConfigMap")
	assert.True(t, ok)
	assert.Equal(t, schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, gvk)

	for _, bad := range []string{"", "no-pipe", "cmp1|", "|v1/ConfigMap", "cmp1|NodePool", "cmp1|v1/"} {
		_, _, ok := decodeComponentKind(bad)
		assert.False(t, ok, bad)
	}
}

func TestManifestKindsRoundTripsThroughPersistence(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}

	first := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	first.Set("cmp-a", nodePoolManifest)
	assert.Len(t, first.DiscoveredGVKs(), 3)
	assert.Len(t, store.ComponentKinds(), 3, "kinds should have been handed to the store")

	restarted := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	assert.Empty(t, restarted.DiscoveredGVKs(), "nothing until it loads")
	restarted.Load()
	assert.ElementsMatch(t, first.DiscoveredGVKs(), restarted.DiscoveredGVKs())
}

func TestPersistDoesNotClobberOtherComponents(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}

	a := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	a.Set("cmp-a", "---\napiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: n\n")
	assert.Len(t, store.ComponentKinds(), 1)

	b := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	b.Set("cmp-b", "---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: c\n")

	assert.Len(t, store.ComponentKinds(), 2, "cmp-a's kind must survive cmp-b's deploy")
	assert.Len(t, b.DiscoveredGVKs(), 2)
}

func TestReleaseOwnershipSurvivesRestart(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}

	first := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	first.SetKinds("cmp-datadog", nil, nil, []string{"datadog-agent"})

	owner, ok := first.ComponentForRelease("datadog-agent")
	assert.True(t, ok)
	assert.Equal(t, "cmp-datadog", owner)

	restarted := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	_, ok = restarted.ComponentForRelease("datadog-agent")
	assert.False(t, ok, "nothing until it loads")

	restarted.Load()
	owner, ok = restarted.ComponentForRelease("datadog-agent")
	assert.True(t, ok, "release ownership must outlive the process")
	assert.Equal(t, "cmp-datadog", owner)
}

func TestReleaseOwnershipReplacedOnReapply(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}
	p := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})

	p.SetKinds("cmp-a", nil, nil, []string{"gone", "kept"})
	p.SetKinds("cmp-a", nil, nil, []string{"kept"})

	_, ok := p.ComponentForRelease("gone")
	assert.False(t, ok, "a release dropped from the module must not stay attributed")
	owner, ok := p.ComponentForRelease("kept")
	assert.True(t, ok)
	assert.Equal(t, "cmp-a", owner)
}

func TestReleaseOwnershipCoexistsWithKindsAndObjects(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}
	p := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})

	p.Set("cmp-chart", nodePoolManifest)
	p.SetKinds("cmp-tf", nil, []string{"ConfigMap//cm"}, []string{"rel"})

	restarted := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})
	restarted.Load()

	assert.Len(t, restarted.DiscoveredGVKs(), 3, "chart kinds survive")
	obj, ok := restarted.ComponentForObject("ConfigMap//cm")
	assert.True(t, ok)
	assert.Equal(t, "cmp-tf", obj)
	rel, ok := restarted.ComponentForRelease("rel")
	assert.True(t, ok)
	assert.Equal(t, "cmp-tf", rel)
}
