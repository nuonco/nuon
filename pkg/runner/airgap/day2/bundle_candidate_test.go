package day2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareBundleContents(t *testing.T) {
	previous := BundleInfo{Contents: []BundleContent{
		{Kind: BundleContentKindComponent, Name: "api", Digest: "sha256:old", ConfigDigest: "sha256:config-old"},
		{Kind: BundleContentKindComponent, Name: "worker", Digest: "sha256:same", ConfigDigest: "sha256:config-same"},
		{Kind: BundleContentKindComponent, Name: "removed", Digest: "sha256:removed"},
	}}
	candidate := BundleInfo{Contents: []BundleContent{
		{Kind: BundleContentKindComponent, Name: "api", Digest: "sha256:new", ConfigDigest: "sha256:config-new"},
		{Kind: BundleContentKindComponent, Name: "worker", Digest: "sha256:same", ConfigDigest: "sha256:config-same"},
		{Kind: BundleContentKindComponent, Name: "added", Digest: "sha256:added"},
	}}

	changes := CompareBundleContents(previous, candidate)
	require.Equal(t, []BundleChange{
		{Kind: BundleContentKindComponent, Name: "added", Change: BundleChangeAdded, CandidateDigest: "sha256:added", PlanStepID: "deploy-added-plan", ApplyStepID: "deploy-added-apply"},
		{Kind: BundleContentKindComponent, Name: "api", Change: BundleChangeChanged, PreviousDigest: "sha256:old", CandidateDigest: "sha256:new", PreviousConfig: "sha256:config-old", CandidateConfig: "sha256:config-new", PlanStepID: "deploy-api-plan", ApplyStepID: "deploy-api-apply"},
		{Kind: BundleContentKindComponent, Name: "removed", Change: BundleChangeRemoved, PreviousDigest: "sha256:removed"},
		{Kind: BundleContentKindComponent, Name: "worker", Change: BundleChangeUnchanged, PreviousDigest: "sha256:same", CandidateDigest: "sha256:same", PreviousConfig: "sha256:config-same", CandidateConfig: "sha256:config-same"},
	}, changes)
}

func TestCompareBundleContentsAddsTerraformSandboxPlanSteps(t *testing.T) {
	changes := CompareBundleContents(
		BundleInfo{Contents: []BundleContent{{Kind: BundleContentKindSandbox, Name: "terraform", Digest: "sha256:old"}}},
		BundleInfo{Contents: []BundleContent{{Kind: BundleContentKindSandbox, Name: "terraform", Digest: "sha256:new"}}},
	)
	require.Len(t, changes, 1)
	require.Equal(t, "sandbox-plan", changes[0].PlanStepID)
	require.Equal(t, "sandbox-apply", changes[0].ApplyStepID)
}

func TestBundlePlanRequestValidation(t *testing.T) {
	valid := Request{
		RefKind: RefKindBundlePlan, RunID: "run-1", BundleDigest: "sha256:candidate",
		CandidateArchiveKey: "deployments/dep/candidate.tar.zst", CandidateRecordKey: "state/day2/candidates/candidate.json",
		PlanStepIDs: []string{"sandbox-plan", "deploy-api-plan"},
	}
	require.NoError(t, valid.ValidateBundlePlan())

	invalid := valid
	invalid.PlanStepIDs = []string{"sandbox-apply"}
	invalid.RunID = "../unsafe"
	require.ErrorContains(t, invalid.ValidateBundlePlan(), "invalid run ID")
	invalid = valid
	invalid.PlanStepIDs = []string{"sandbox-plan", "sandbox-plan"}
	require.ErrorContains(t, invalid.ValidateBundlePlan(), "duplicate")
}

func TestCompareBundleContentsCarriesActionDefinitions(t *testing.T) {
	previousDefinition := &BundleActionDefinition{Steps: []BundleActionStep{{Name: "restart", Command: "kubectl rollout restart"}}}
	candidateDefinition := &BundleActionDefinition{Steps: []BundleActionStep{{Name: "restart", Command: "kubectl rollout restart --timeout=5m"}}}
	changes := CompareBundleContents(
		BundleInfo{Contents: []BundleContent{{Kind: BundleContentKindAction, Name: "restart", ConfigDigest: "sha256:old", ActionDefinition: previousDefinition}}},
		BundleInfo{Contents: []BundleContent{{Kind: BundleContentKindAction, Name: "restart", ConfigDigest: "sha256:new", ActionDefinition: candidateDefinition}}},
	)

	require.Len(t, changes, 1)
	require.Equal(t, BundleChangeChanged, changes[0].Change)
	require.Equal(t, previousDefinition, changes[0].PreviousActionDefinition)
	require.Equal(t, candidateDefinition, changes[0].CandidateActionDefinition)
}
