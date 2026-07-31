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

// Kubernetes keeps events ~1h. A resource that recovered after the event fired
// must not stay degraded until the event expires.
func TestStaleWarningDoesNotPinRecoveredResource(t *testing.T) {
	firedAt := time.Now().Add(-4 * time.Minute)
	recoveredAt := firedAt.Add(time.Minute).UTC().Format(time.RFC3339)

	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: firedAt}
	res := resourceModel(issuerGVR, issuer("True", recoveredAt), warn)

	assert.Equal(t, healthHealthy, res.Health, "the condition transitioned after the event fired")
	assert.Empty(t, res.Message)
	assert.NotContains(t, res.Details, "ErrGetKeyPair", "a retired warning is not a diagnosis")
}

func TestWarningStillCountsWhenNewerThanRecovery(t *testing.T) {
	transitioned := time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339)
	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: time.Now().Add(-time.Minute)}

	res := resourceModel(issuerGVR, issuer("True", transitioned), warn)

	assert.Equal(t, healthDegraded, res.Health, "the event is the newer evidence")
	assert.Contains(t, res.Message, "ErrGetKeyPair")
}

func TestWarningCountsWhenResourceIsNotReady(t *testing.T) {
	recent := time.Now().UTC().Format(time.RFC3339)
	warn := &warningEvent{reason: "ErrGetKeyPair", message: "secret not found", at: time.Now().Add(-time.Minute)}

	// Ready=False must never be superseded, whatever the timestamps say.
	res := resourceModel(issuerGVR, issuer("False", recent), warn)
	assert.Equal(t, healthDegraded, res.Health)
}

// An event with no usable timestamp keeps the old, conservative behaviour.
func TestWarningWithoutTimestampStillCounts(t *testing.T) {
	recovered := time.Now().UTC().Format(time.RFC3339)
	warn := &warningEvent{reason: "Failed", message: "boom"}

	res := resourceModel(issuerGVR, issuer("True", recovered), warn)
	assert.Equal(t, healthDegraded, res.Health)
}
