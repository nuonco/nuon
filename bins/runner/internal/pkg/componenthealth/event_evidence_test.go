package componenthealth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func issuer(readyStatus, transitionedAt string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Issuer",
		"metadata":   map[string]any{"name": "iss", "namespace": "whoami"},
		"status": map[string]any{"conditions": []any{
			map[string]any{"type": "Ready", "status": readyStatus, "lastTransitionTime": transitionedAt},
		}},
	}}
}

var issuerGVR = schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}

// The invariant this package exists to hold: a Warning event may explain a
// verdict but can never set one.
//
// Events are edge-triggered and have no "all clear", so any code that lets one
// decide health must invent an expiry, and every expiry is a fabricated claim
// about how long the fault lasted. That is how a one-minute HPA metrics gap
// pinned a component degraded for exactly 15m + 2 reports.
//
// Feeding the same object every shape of event history and asserting one
// verdict is what mechanically forbids reintroducing it.
func TestEventNeverChangesVerdict(t *testing.T) {
	t.Parallel()

	now := time.Now()
	histories := []*warningEvent{
		nil,
		{reason: "ErrGetKeyPair", message: "secret not found"},
		{reason: "ErrGetKeyPair", message: "secret not found", at: now},
		{reason: "ErrGetKeyPair", message: "secret not found", at: now.Add(-time.Second)},
		{reason: "ErrGetKeyPair", message: "secret not found", at: now.Add(-14 * time.Minute)},
		{reason: "ErrGetKeyPair", message: "secret not found", at: now.Add(-time.Hour)},
	}

	recent := now.UTC().Format(time.RFC3339)
	old := now.Add(-time.Hour).UTC().Format(time.RFC3339)

	objects := map[string]*unstructured.Unstructured{
		"ready issuer":                 issuer("True", recent),
		"ready issuer, old transition": issuer("True", old),
		"failing issuer":               issuer("False", recent),
		"hpa waiting on metrics":       hpaObj("kitchen-sink-api", hpaCond("AbleToScale", "True", "ReadyForNewScale", "recommended size matches current size"), hpaCond("ScalingActive", "False", "FailedGetResourceMetric", "did not receive metrics for targeted pods")),
		"hpa scaling normally":         hpaObj("kitchen-sink-ui", hpaCond("AbleToScale", "True", "ReadyForNewScale", "recommended size matches current size"), hpaCond("ScalingActive", "True", "ValidMetricFound", "the HPA was able to compute the replica count")),
	}

	for name, obj := range objects {
		gvr := issuerGVR
		if obj.GetKind() == "HorizontalPodAutoscaler" {
			gvr = hpaGVR
		}
		want := resourceModel(gvr, obj, nil, "").Health

		for _, warn := range histories {
			got := resourceModel(gvr, obj, warn, "")
			assert.Equal(t, want, got.Health,
				"%s: an event changed the verdict — health must be a pure function of the object's own status", name)
		}
	}
}

// A healthy resource must carry no trace of an event: that is the whole failure
// mode — a recovered object rendered with a stale message pasted over it.
func TestHealthyResourceDiscardsEvent(t *testing.T) {
	t.Parallel()

	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: time.Now()}
	res := resourceModel(issuerGVR, issuer("True", time.Now().UTC().Format(time.RFC3339)), warn, "")

	assert.Equal(t, healthHealthy, res.Health)
	assert.Empty(t, res.Message)
	assert.NotContains(t, res.Details, "ErrGetKeyPair")
}

// The diagnostic value of events is kept: a resource its own status reports as
// failing still gets the event, which is usually the only place the cause is
// written down.
func TestFailingResourceKeepsEventAsEvidence(t *testing.T) {
	t.Parallel()

	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: time.Now()}
	res := resourceModel(issuerGVR, issuer("False", time.Now().UTC().Format(time.RFC3339)), warn, "")

	assert.Equal(t, healthDegraded, res.Health)
	assert.Contains(t, res.Details, "ErrGetKeyPair", "the cause must still reach the detail view")
}

// An event fills the headline only when status left it blank, so it can add to
// what the object says but never contradict it.
func TestEventDoesNotOverwriteStatusMessage(t *testing.T) {
	t.Parallel()

	obj := issuer("False", time.Now().UTC().Format(time.RFC3339))
	conds, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	conds[0].(map[string]any)["message"] = "the CA is unreachable"
	_ = unstructured.SetNestedSlice(obj.Object, conds, "status", "conditions")

	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: time.Now()}
	res := resourceModel(issuerGVR, obj, warn, "")

	assert.Equal(t, "the CA is unreachable", res.Message)
}

// A controller has not necessarily read the spec it is being graded against.
func TestStaleGenerationReadsProgressing(t *testing.T) {
	t.Parallel()

	obj := issuer("True", time.Now().UTC().Format(time.RFC3339))
	obj.SetGeneration(4)
	mustSet(t, unstructured.SetNestedField(obj.Object, int64(3), "status", "observedGeneration"))

	health, _, _ := assessResource(obj)
	assert.Equal(t, healthProgressing, health,
		"conditions from generation 3 say nothing about generation 4")

	mustSet(t, unstructured.SetNestedField(obj.Object, int64(4), "status", "observedGeneration"))
	health, _, _ = assessResource(obj)
	assert.Equal(t, healthHealthy, health)
}

// Most kinds never write observedGeneration; claiming they are all mid-rollout
// forever would take every CRD permanently progressing.
func TestMissingObservedGenerationIsNotStale(t *testing.T) {
	t.Parallel()

	obj := issuer("True", time.Now().UTC().Format(time.RFC3339))
	obj.SetGeneration(7)

	assert.False(t, staleGeneration(obj))

	mustSet(t, unstructured.SetNestedField(obj.Object, int64(0), "status", "observedGeneration"))
	assert.False(t, staleGeneration(obj), "a zero value is not evidence the controller fell behind")
}

func mustSet(t *testing.T, err error) {
	t.Helper()
	assert.NoError(t, err)
}
