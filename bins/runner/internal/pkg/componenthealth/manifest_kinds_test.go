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
