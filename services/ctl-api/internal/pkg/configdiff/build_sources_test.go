package configdiff

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type buildSourceFact struct {
	evidence   string
	config     int
	content    int
	provenance int
	checksum   string
	digest     string
}

var buildEvidenceKinds = []string{
	"legacy", "complete", "unknown", "unsupported", "uncaptured", "corrupt",
	"missing_config", "invalid_path", "invalid_commit", "complete_with_reasons",
	"missing_checksum", "bare_checksum",
}

func (f buildSourceFact) identity() (*sourceidentity.Identity, string) {
	if f.evidence == "legacy" {
		return nil, ""
	}
	if f.evidence == "bare_checksum" {
		return nil, "bare-source-checksum"
	}
	i := sourceidentity.New("terraform", fmt.Sprintf("config-%d", f.config))
	i.Captured = true
	i.Commit = strings.Repeat(fmt.Sprintf("%x", f.provenance+1), 40)
	i.Repository = fmt.Sprintf("https://example.com/acme/repo-%d", f.provenance)
	i.Files = []sourceidentity.File{{Path: "main.tf", SHA256: strings.Repeat(fmt.Sprintf("%x", f.content+1), 64)}}
	switch f.evidence {
	case "unknown":
		i.Coverage, i.Reasons = sourceidentity.Unknown, []string{"uninspected_dependency"}
	case "unsupported":
		i.Version++
	case "uncaptured":
		i.Captured = false
	case "missing_config":
		i.ConfigChecksum = ""
	case "invalid_path":
		i.Files[0].Path = "../main.tf"
	case "invalid_commit":
		i.Commit = "main"
	case "complete_with_reasons":
		i.Reasons = []string{"uninspected_dependency"}
	}
	i.Seal()
	sum := i.Checksum
	switch f.evidence {
	case "corrupt":
		i.Files[0].Executable = true
	case "missing_checksum":
		sum = ""
	}
	return i, sum
}

func (f buildSourceFact) component(id string) *app.ComponentBuild {
	identity, checksum := f.identity()
	return &app.ComponentBuild{
		ID: id, SourceIdentity: identity, SourceChecksum: checksum,
		SourceDigest: f.digest, Checksum: f.checksum,
		SourceRef:   fmt.Sprintf("example.com/acme/image:tag-%d", f.provenance),
		SourceImage: fmt.Sprintf("example.com/acme/image-%d", f.provenance),
	}
}

func (f buildSourceFact) sandbox(id string) *app.AppSandboxBuild {
	identity, checksum := f.identity()
	return &app.AppSandboxBuild{ID: id, SourceIdentity: identity, SourceChecksum: checksum, Checksum: f.checksum}
}

func sourceFactsEqual(old, next buildSourceFact, legacy bool) bool {
	if old.evidence == "complete" && next.evidence == "complete" && old.config == next.config && old.content == next.content {
		return true
	}
	oldWithoutIdentity := old.evidence == "legacy" || old.evidence == "bare_checksum"
	nextWithoutIdentity := next.evidence == "legacy" || next.evidence == "bare_checksum"
	return legacy && oldWithoutIdentity && nextWithoutIdentity && old.digest != "" && old.digest == next.digest
}

func TestBuildSourcesEvidenceProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		old := buildSourceFact{
			config: hegel.Draw(t, hegel.Integers(0, 2)), content: hegel.Draw(t, hegel.Integers(0, 2)),
			provenance: hegel.Draw(t, hegel.Integers(0, 2)),
			checksum:   hegel.Draw(t, hegel.SampledFrom([]string{"", "config-a", "config-b"})),
			digest:     hegel.Draw(t, hegel.SampledFrom([]string{"", "sha256:" + strings.Repeat("a", 64), "sha256:" + strings.Repeat("b", 64)})),
		}
		next := buildSourceFact{
			config: hegel.Draw(t, hegel.Integers(0, 2)), content: hegel.Draw(t, hegel.Integers(0, 2)),
			provenance: hegel.Draw(t, hegel.Integers(0, 2)),
			checksum:   hegel.Draw(t, hegel.SampledFrom([]string{"", "config-a", "config-b"})),
			digest:     hegel.Draw(t, hegel.SampledFrom([]string{"", "sha256:" + strings.Repeat("a", 64), "sha256:" + strings.Repeat("b", 64)})),
		}
		components := make([][2]*app.ComponentBuild, len(buildEvidenceKinds))
		sandboxes := make([][2]*app.AppSandboxBuild, len(buildEvidenceKinds))
		for i, kind := range buildEvidenceKinds {
			old.evidence, next.evidence = kind, kind
			components[i] = [2]*app.ComponentBuild{old.component(""), next.component("")}
			sandboxes[i] = [2]*app.AppSandboxBuild{old.sandbox(""), next.sandbox("")}
		}
		for i, oldKind := range buildEvidenceKinds {
			for j, nextKind := range buildEvidenceKinds {
				old.evidence, next.evidence = oldKind, nextKind
				a, b := components[i][0], components[j][1]
				x, y := sandboxes[i][0], sandboxes[j][1]
				for _, ids := range [][2]string{{"build-old", "build-new"}, {"", ""}, {"", "build-new"}, {"build-old", ""}} {
					a.ID, b.ID, x.ID, y.ID = ids[0], ids[1], ids[0], ids[1]
					wantComponent, wantSandbox := sourceFactsEqual(old, next, true), sourceFactsEqual(old, next, false)
					require.Equal(t, wantComponent, ComponentSourcesEqual(a, b), "%s/%s, ids=%v", oldKind, nextKind, ids)
					require.Equal(t, wantComponent, ComponentSourcesEqual(b, a))
					require.Equal(t, wantSandbox, SandboxSourcesEqual(x, y), "%s/%s, ids=%v", oldKind, nextKind, ids)
					require.Equal(t, wantSandbox, SandboxSourcesEqual(y, x))
				}
			}
			old.evidence = oldKind
			require.Equal(t, sourceFactsEqual(old, old, true), ComponentSourcesEqual(old.component("old"), old.component("new")), "matching evidence: %s", oldKind)
			require.Equal(t, sourceFactsEqual(old, old, false), SandboxSourcesEqual(old.sandbox("old"), old.sandbox("new")), "matching evidence: %s", oldKind)
			component, sandbox := old.component("immutable-build"), old.sandbox("immutable-build")
			componentCopy, sandboxCopy := *component, *sandbox
			require.True(t, ComponentSourcesEqual(component, &componentCopy), oldKind)
			require.True(t, SandboxSourcesEqual(sandbox, &sandboxCopy), oldKind)
			require.False(t, ComponentSourcesEqual(nil, component))
			require.False(t, ComponentSourcesEqual(component, nil))
			require.False(t, SandboxSourcesEqual(nil, sandbox))
			require.False(t, SandboxSourcesEqual(sandbox, nil))
		}
		require.False(t, ComponentSourcesEqual(nil, nil))
		require.False(t, SandboxSourcesEqual(nil, nil))
	})
}

func TestBuildSourcesProvenanceAndArtifactProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		old := buildSourceFact{evidence: "complete", config: hegel.Draw(t, hegel.Integers(0, 2)), content: hegel.Draw(t, hegel.Integers(0, 2))}
		next := old
		next.provenance = 1
		next.checksum = "different-build-checksum"
		next.digest = "sha256:" + strings.Repeat("b", 64)
		artifact := app.RunnerJob{Execution: &app.RunnerJobExecution{Outputs: &app.RunnerJobExecutionOutputs{Outputs: []byte(`{"artifact_digest":"sha256:` + strings.Repeat("a", 64) + `"}`)}}}
		otherArtifact := app.RunnerJob{Execution: &app.RunnerJobExecution{Outputs: &app.RunnerJobExecutionOutputs{Outputs: []byte(`{"artifact_digest":"sha256:` + strings.Repeat("b", 64) + `"}`)}}}
		a, b := old.component("old"), next.component("new")
		x, y := old.sandbox("old"), next.sandbox("new")
		a.RunnerJob, b.RunnerJob, x.RunnerJob, y.RunnerJob = artifact, otherArtifact, artifact, otherArtifact
		require.True(t, ComponentSourcesEqual(a, b))
		require.True(t, SandboxSourcesEqual(x, y))
		next = old
		next.content = (old.content + 1) % 3
		b, y = next.component("new"), next.sandbox("new")
		b.RunnerJob, y.RunnerJob = artifact, artifact
		require.False(t, ComponentSourcesEqual(a, b))
		require.False(t, SandboxSourcesEqual(x, y))
		next = old
		next.config = (old.config + 1) % 3
		require.False(t, ComponentSourcesEqual(old.component("old"), next.component("new")))
		require.False(t, SandboxSourcesEqual(old.sandbox("old"), next.sandbox("new")))
	})
}

func TestBuildSourcesRejectMismatchedChecksumProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fact := buildSourceFact{evidence: "complete", config: hegel.Draw(t, hegel.Integers(0, 2)), content: hegel.Draw(t, hegel.Integers(0, 2))}
		for _, corruptOld := range []bool{false, true} {
			old, next := fact.component("old"), fact.component("new")
			a, b := fact.sandbox("old"), fact.sandbox("new")
			if corruptOld {
				old.SourceChecksum, a.SourceChecksum = "unvalidated", "unvalidated"
			} else {
				next.SourceChecksum, b.SourceChecksum = "unvalidated", "unvalidated"
			}
			require.False(t, ComponentSourcesEqual(old, next))
			require.False(t, SandboxSourcesEqual(a, b))
			old.SourceChecksum, next.SourceChecksum = "unvalidated", "unvalidated"
			a.SourceChecksum, b.SourceChecksum = "unvalidated", "unvalidated"
			require.False(t, ComponentSourcesEqual(old, next))
			require.False(t, SandboxSourcesEqual(a, b))
		}
	})
}

func TestSourceChangedRootsProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		old := &configBuildSources{components: map[string]*app.ComponentBuild{}}
		next := &configBuildSources{components: map[string]*app.ComponentBuild{}}
		cfg := &app.AppConfig{}
		var want []pkgdiff.NodeID
		for i := range hegel.Draw(t, hegel.Integers(2, 6)) {
			name := fmt.Sprintf("component-%d", i)
			cfg.ComponentConfigConnections = append(cfg.ComponentConfigConnections, app.ComponentConfigConnection{ComponentID: name, ComponentName: name})
			fact := buildSourceFact{evidence: hegel.Draw(t, hegel.SampledFrom(buildEvidenceKinds))}
			other := fact
			other.content = hegel.Draw(t, hegel.Integers(0, 1))
			old.components[name], next.components[name] = fact.component(name+"-old"), other.component(name+"-new")
			if !sourceFactsEqual(fact, other, true) {
				want = append(want, pkgdiff.NodeID("component."+name))
			}
		}
		fact := buildSourceFact{evidence: hegel.Draw(t, hegel.SampledFrom(buildEvidenceKinds))}
		other := fact
		other.content = hegel.Draw(t, hegel.Integers(0, 1))
		old.sandbox, next.sandbox = fact.sandbox("sandbox-old"), other.sandbox("sandbox-new")
		if !sourceFactsEqual(fact, other, false) {
			want = append(want, pkgdiff.NodeID("sandbox"))
		}
		require.Equal(t, want, sourceChangedRoots(old, next, cfg))
	})
}

func TestLegacyDigestCompatibilityProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		digest := "sha256:" + strings.Repeat(hegel.Draw(t, hegel.SampledFrom([]string{"a", "b", "c"})), 64)
		fact := buildSourceFact{evidence: "legacy", digest: digest}
		old, next := fact.component("old"), fact.component("new")
		next.Checksum = "different-config-checksum"
		require.True(t, ComponentSourcesEqual(old, next))
		require.False(t, SandboxSourcesEqual(fact.sandbox("old"), fact.sandbox("new")))
		for _, unsupported := range []string{"", "image:latest", "sha512:" + strings.Repeat("a", 128)} {
			old.SourceDigest, next.SourceDigest = unsupported, unsupported
			require.False(t, ComponentSourcesEqual(old, next))
		}
	})
}

func TestPersistedDependencyImpactsModelProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		n := hegel.Draw(t, hegel.Integers(3, 6))
		node := func(i int) pkgdiff.NodeID {
			if i == n {
				return "sandbox"
			}
			return pkgdiff.NodeID(fmt.Sprintf("component.resource-%d", i))
		}
		type edgeFact struct {
			consumer, producer int
			reason             pkgdiff.EdgeReason
		}
		edges := map[edgeFact]bool{}
		configs := []*app.AppConfig{{}, {}}
		if hegel.Draw(t, hegel.Booleans()) {
			configs[0] = nil
		}
		for _, cfg := range configs {
			if cfg == nil {
				continue
			}
			for i := range n {
				conn := app.ComponentConfigConnection{ComponentID: fmt.Sprintf("cmp-%d", i), ComponentName: fmt.Sprintf("resource-%d", i)}
				for j := range i {
					if hegel.Draw(t, hegel.Booleans()) {
						conn.ComponentDependencyIDs = append(conn.ComponentDependencyIDs, fmt.Sprintf("cmp-%d", j))
						edges[edgeFact{i, j, "component_dependency"}] = true
					}
					if hegel.Draw(t, hegel.Booleans()) {
						conn.References = append(conn.References, fmt.Sprintf("component|resource-%d|output|", j))
						edges[edgeFact{i, j, "component_reference"}] = true
					}
				}
				if hegel.Draw(t, hegel.Booleans()) {
					conn.References = append(conn.References, "sandbox||output|")
					edges[edgeFact{i, n, "sandbox_output"}] = true
				}
				cfg.ComponentConfigConnections = append(cfg.ComponentConfigConnections, conn)
			}
		}
		direct, reached := make([]bool, n+1), make([]bool, n+1)
		var roots []pkgdiff.NodeID
		for i := range n + 1 {
			direct[i] = hegel.Draw(t, hegel.Booleans())
			reached[i] = direct[i]
			if direct[i] {
				roots = append(roots, node(i))
			}
		}
		for range n + 1 {
			for edge := range edges {
				if reached[edge.producer] {
					reached[edge.consumer] = true
				}
			}
		}
		want := map[pkgdiff.NodeID][]pkgdiff.ImpactReason{}
		for edge := range edges {
			if reached[edge.producer] && !direct[edge.consumer] {
				want[node(edge.consumer)] = append(want[node(edge.consumer)], pkgdiff.ImpactReason{From: node(edge.producer), Edge: edge.reason})
			}
		}
		got := persistedDependencyImpacts(configs[0], configs[1], roots)
		require.Len(t, got, len(want))
		for id, reasons := range want {
			require.ElementsMatch(t, reasons, got[id], string(id))
		}
		require.Equal(t, got, persistedDependencyImpacts(configs[0], configs[1], roots), "extraction and propagation must be repeatable")
	})
}
