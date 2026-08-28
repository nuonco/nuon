package customermanaged

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
)

var _ ociarchive.Source = (*BundleSource)(nil)

// BundleSource resolves sync-oci plan sources to artifacts packaged in an
// customer-managed bundle. Plan envelopes are exported from an online reference
// install, so every sync-oci step still names the vendor's management ECR as
// its source with a build ID as the tag. The bundle's provenance records the
// build ID each packaged member was resolved from, which lets an offline run
// map a source tag back to the bundled artifact without any registry access.
type BundleSource struct {
	mu    sync.RWMutex
	byTag map[string]bundleArtifact
}

type bundleArtifact struct {
	store  oras.ReadOnlyTarget
	digest digest.Digest
}

func NewBundleSource(store oras.ReadOnlyTarget, members []bundle.Member, provenance json.RawMessage) (*BundleSource, error) {
	var doc struct {
		BuildIDs map[string]string `json:"build_ids"`
	}
	if len(provenance) > 0 {
		if err := json.Unmarshal(provenance, &doc); err != nil {
			return nil, fmt.Errorf("decode bundle provenance: %w", err)
		}
	}
	byKey := make(map[string]digest.Digest, len(members))
	for _, member := range members {
		byKey[member.Key] = member.Digest
	}
	byTag := make(map[string]bundleArtifact, len(doc.BuildIDs))
	for key, buildID := range doc.BuildIDs {
		d, ok := byKey[key]
		if !ok {
			return nil, fmt.Errorf("bundle provenance references member %q which is not in the bundle manifest", key)
		}
		byTag[buildID] = bundleArtifact{store: store, digest: d}
	}
	return &BundleSource{byTag: byTag}, nil
}

func (s *BundleSource) AddPlanAliases(envelope *Envelope) error {
	aliases := make(map[string]bundleArtifact)
	for _, step := range envelope.Steps {
		var cp plantypes.CompositePlan
		if err := json.Unmarshal(step.CompositePlan, &cp); err != nil {
			return fmt.Errorf("decode plan step %s: %w", step.ID, err)
		}
		if cp.SyncOCIPlan == nil {
			continue
		}
		s.mu.RLock()
		artifact, found := s.byTag[cp.SyncOCIPlan.SrcTag]
		s.mu.RUnlock()
		if !found {
			return fmt.Errorf("step %s source tag %s is not packaged in the bundle", step.ID, cp.SyncOCIPlan.SrcTag)
		}
		aliases[cp.SyncOCIPlan.DstTag] = artifact
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for tag, artifact := range aliases {
		s.byTag[tag] = artifact
	}
	return nil
}

// Merge returns a source containing both bundles without mutating either
// source. Candidate tags take precedence when the same build ID is present.
func (s *BundleSource) Merge(candidate *BundleSource) *BundleSource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	candidate.mu.RLock()
	defer candidate.mu.RUnlock()
	merged := &BundleSource{byTag: make(map[string]bundleArtifact, len(s.byTag)+len(candidate.byTag))}
	for tag, artifact := range s.byTag {
		merged.byTag[tag] = artifact
	}
	for tag, artifact := range candidate.byTag {
		merged.byTag[tag] = artifact
	}
	return merged
}

func (s *BundleSource) Overlay(candidate *BundleSource) func() {
	s.mu.Lock()
	candidate.mu.RLock()
	previous := make(map[string]bundleArtifact, len(candidate.byTag))
	missing := make(map[string]bool, len(candidate.byTag))
	for tag, artifact := range candidate.byTag {
		old, found := s.byTag[tag]
		if found {
			previous[tag] = old
		} else {
			missing[tag] = true
		}
		s.byTag[tag] = artifact
	}
	candidate.mu.RUnlock()
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for tag, artifact := range previous {
			s.byTag[tag] = artifact
		}
		for tag := range missing {
			delete(s.byTag, tag)
		}
	}
}

// ResolveArchive maps a plan's OCI source tag to the packaged artifact in the
// bundle store, implementing ociarchive.Source for sandbox source unpacks.
func (s *BundleSource) ResolveArchive(tag string) (oras.ReadOnlyTarget, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	artifact, ok := s.byTag[tag]
	if !ok {
		return nil, "", false
	}
	return artifact.store, artifact.digest.String(), true
}

// MissingPlanSources returns a description of every plan source in the
// envelope that cannot be served from the bundle: sync-oci source tags and
// sandbox run sources (which must be OCI-backed; git sources cannot be
// packaged). A non-empty result means the run would need network access and
// must not start.
func (s *BundleSource) MissingPlanSources(envelope *Envelope) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var missing []string
	for _, step := range envelope.Steps {
		var cp plantypes.CompositePlan
		if err := json.Unmarshal(step.CompositePlan, &cp); err != nil {
			continue
		}
		if plan := cp.SyncOCIPlan; plan != nil {
			if _, ok := s.byTag[plan.SrcTag]; !ok {
				missing = append(missing, fmt.Sprintf("step %s: %s:%s", step.ID, plan.Src.Repository, plan.SrcTag))
			}
		}
		if plan := cp.SandboxRunPlan; plan != nil {
			switch {
			case plan.OCISource != nil:
				if _, ok := s.byTag[plan.OCISource.Tag]; !ok {
					missing = append(missing, fmt.Sprintf("step %s: sandbox source tag %s", step.ID, plan.OCISource.Tag))
				}
			case plan.GitSource != nil:
				missing = append(missing, fmt.Sprintf("step %s: sandbox source is git-backed (%s); offline plans must use a packaged OCI source", step.ID, plan.GitSource.URL))
			}
		}
	}
	sort.Strings(missing)
	return missing
}

// Copier wraps the standard OCI copier so remote-source copies are served
// from the bundle instead. Sources that are not packaged fail hard: an
// offline run must never fall back to the network.
func (s *BundleSource) Copier(inner ocicopy.Copier) ocicopy.Copier {
	return &bundleCopier{inner: inner, source: s}
}

type bundleCopier struct {
	inner  ocicopy.Copier
	source *BundleSource
}

var _ ocicopy.Copier = (*bundleCopier)(nil)

func (c *bundleCopier) Copy(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	c.source.mu.RLock()
	artifact, ok := c.source.byTag[srcTag]
	c.source.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("source %s:%s is not packaged in the bundle; offline runs cannot pull from remote registries", srcCfg.Repository, srcTag)
	}
	return c.inner.CopyFromStore(ctx, artifact.store, artifact.digest.String(), dstCfg, dstTag)
}

func (c *bundleCopier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	return c.inner.CopyFromStore(ctx, store, srcTag, dstCfg, dstTag)
}

func (c *bundleCopier) CopyFromLocalRegistry(ctx context.Context, localTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	return c.inner.CopyFromLocalRegistry(ctx, localTag, dstCfg, dstTag)
}
