// Package runnerstatussignals defines the signal contract used to wake
// org provision/reprovision workflows when the runner they are waiting
// on changes status.
//
// The org Provision and Reprovision workflows poll runner status every
// 10s waiting for the runner to become Active. This package replaces
// that poll loop with a push signal, following the same pattern used
// for ProcessJob in processjobsignals.
package runnerstatussignals

import "fmt"

// Namespace is the Temporal namespace for org workflows.
const Namespace = "orgs"

// SignalName is the wake-up signal sent to org provision/reprovision workflows.
const SignalName = "runner_status_wakeup"

const ReasonStatusChanged = "runner_status_changed"

// WakeUp is the signal payload. The workflow re-reads runner status
// authoritatively on any wake-up; the reason is for observability only.
type WakeUp struct {
	Reason string `json:"reason"`
}

// ProvisionID returns the deterministic workflow ID for an org's Provision workflow.
func ProvisionID(orgID string) string { return fmt.Sprintf("provision-%s", orgID) }

// ReprovisionID returns the deterministic workflow ID for an org's Reprovision workflow.
func ReprovisionID(orgID string) string { return fmt.Sprintf("reprovision-%s", orgID) }
