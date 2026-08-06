package airgap

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
)

type recordingCopier struct {
	store  oras.ReadOnlyTarget
	srcTag string
	dstTag string
}

var _ ocicopy.Copier = (*recordingCopier)(nil)

func (r *recordingCopier) Copy(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	panic("bundle copier must never fall back to a remote source copy")
}

func (r *recordingCopier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	r.store, r.srcTag, r.dstTag = store, srcTag, dstTag
	return &ocispec.Descriptor{}, nil
}

func (r *recordingCopier) CopyFromLocalRegistry(ctx context.Context, localTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	return &ocispec.Descriptor{}, nil
}

func testBundleSource(t *testing.T) (*BundleSource, digest.Digest, oras.ReadOnlyTarget) {
	t.Helper()
	store := memory.New()
	d := digest.FromString("whoami-build")
	members := []bundle.Member{
		{Key: "component:whoami", Kind: "component", Name: "whoami", Digest: d},
		{Key: "sandbox", Kind: "sandbox", Name: "terraform", Digest: digest.FromString("sandbox-build")},
	}
	provenance := json.RawMessage(`{"app_config_id":"appcfg","build_ids":{"component:whoami":"bldwhoami","sandbox":"bldsandbox"}}`)
	source, err := NewBundleSource(store, members, provenance)
	require.NoError(t, err)
	return source, d, store
}

func TestBundleCopierServesPackagedSourceFromStore(t *testing.T) {
	source, d, store := testBundleSource(t)
	inner := &recordingCopier{}
	copier := source.Copier(inner)

	src := &configs.OCIRegistryRepository{Repository: "org/app"}
	dst := &configs.OCIRegistryRepository{Repository: "customer/install"}
	_, err := copier.Copy(context.Background(), src, "bldwhoami", dst, "dpl123")
	require.NoError(t, err)
	require.Equal(t, store, inner.store)
	require.Equal(t, d.String(), inner.srcTag)
	require.Equal(t, "dpl123", inner.dstTag)
}

func TestBundleCopierRejectsUnpackagedSource(t *testing.T) {
	source, _, _ := testBundleSource(t)
	copier := source.Copier(&recordingCopier{})

	src := &configs.OCIRegistryRepository{Repository: "org/app"}
	dst := &configs.OCIRegistryRepository{Repository: "customer/install"}
	_, err := copier.Copy(context.Background(), src, "bldunknown", dst, "dpl123")
	require.ErrorContains(t, err, "not packaged in the bundle")
	require.ErrorContains(t, err, "org/app:bldunknown")
}

func TestNewBundleSourceRejectsProvenanceWithMissingMember(t *testing.T) {
	provenance := json.RawMessage(`{"build_ids":{"component:ghost":"bldghost"}}`)
	_, err := NewBundleSource(memory.New(), nil, provenance)
	require.ErrorContains(t, err, `component:ghost`)
}

func TestNewBundleSourceRejectsMalformedProvenance(t *testing.T) {
	_, err := NewBundleSource(memory.New(), nil, json.RawMessage(`{`))
	require.ErrorContains(t, err, "decode bundle provenance")
}

func syncStep(t *testing.T, id, srcTag string) Step {
	t.Helper()
	cp := plantypes.CompositePlan{SyncOCIPlan: &plantypes.SyncOCIPlan{
		Src:    &configs.OCIRegistryRepository{Repository: "org/app"},
		SrcTag: srcTag,
		Dst:    &configs.OCIRegistryRepository{Repository: "customer/install"},
		DstTag: "dst-" + id,
	}}
	raw, err := json.Marshal(cp)
	require.NoError(t, err)
	return Step{ID: id, JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync", CompositePlan: raw}
}

func sandboxStep(t *testing.T, id string, gitSource *plantypes.GitSource, ociSource *plantypes.OCISource) Step {
	t.Helper()
	cp := plantypes.CompositePlan{SandboxRunPlan: &plantypes.SandboxRunPlan{GitSource: gitSource, OCISource: ociSource}}
	raw, err := json.Marshal(cp)
	require.NoError(t, err)
	return Step{ID: id, JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: raw}
}

func TestMissingPlanSources(t *testing.T) {
	source, _, _ := testBundleSource(t)
	envelope := &Envelope{Steps: []Step{
		syncStep(t, "job1", "bldwhoami"),
		syncStep(t, "job2", "bldmissing"),
		{ID: "job3", JobType: "sandbox-terraform", CompositePlan: json.RawMessage(`{}`)},
	}}
	missing := source.MissingPlanSources(envelope)
	require.Equal(t, []string{"step job2: org/app:bldmissing"}, missing)

	envelope.Steps = envelope.Steps[:1]
	require.Empty(t, source.MissingPlanSources(envelope))
}

func TestMissingPlanSourcesSandbox(t *testing.T) {
	source, _, _ := testBundleSource(t)

	packaged := sandboxStep(t, "job1", nil, &plantypes.OCISource{Tag: "bldsandbox"})
	require.Empty(t, source.MissingPlanSources(&Envelope{Steps: []Step{packaged}}))

	unpackaged := sandboxStep(t, "job2", nil, &plantypes.OCISource{Tag: "bldother"})
	missing := source.MissingPlanSources(&Envelope{Steps: []Step{unpackaged}})
	require.Equal(t, []string{"step job2: sandbox source tag bldother"}, missing)

	gitBacked := sandboxStep(t, "job3", &plantypes.GitSource{URL: "https://github.com/nuonco/aws-eks-auto-sandbox", Ref: "main", Path: "."}, nil)
	missing = source.MissingPlanSources(&Envelope{Steps: []Step{gitBacked}})
	require.Len(t, missing, 1)
	require.Contains(t, missing[0], "git-backed")
	require.Contains(t, missing[0], "github.com/nuonco/aws-eks-auto-sandbox")
}

func TestResolveArchive(t *testing.T) {
	source, _, store := testBundleSource(t)

	gotStore, ref, ok := source.ResolveArchive("bldsandbox")
	require.True(t, ok)
	require.Equal(t, store, gotStore)
	require.Equal(t, digest.FromString("sandbox-build").String(), ref)

	_, _, ok = source.ResolveArchive("bldunknown")
	require.False(t, ok)
}
