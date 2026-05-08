// Mirror of the Go SubOps map in services/ctl-api/internal/pkg/interests/types.go.
// Empty `ops: []` on a ResourceCfg means "all sub-ops" — we render this list to
// let the user narrow the per-resource filter, never to enumerate the wire
// payload. Keep this in sync with the Go map; the backend validator rejects
// unknown ops outright.
//
// Note: "drift" is intentionally NOT listed for components or sandboxes — the
// backend matcher unconditionally suppresses drift workflow lifecycle events.
// Drift surfaces only through the dedicated drift_detected event class, gated
// by the per-resource drift_detected toggle (see RESOURCES_WITH_DRIFT_DETECTED).

import type { ResourceKind } from './types'

export const SUB_OPS: Record<ResourceKind, string[]> = {
  installs: ['provision', 'deprovision', 'reprovision'],
  components: ['deploy', 'teardown'],
  sandboxes: ['provision', 'reprovision', 'deprovision'],
  install_configurations: ['inputs', 'secrets'],
  runners: ['provision', 'reprovision', 'inactive'],
  actions: ['run'],
}

// Resources whose workflows can produce a drift_detected event. Mirrors the Go
// SupportsDriftDetected helper. The picker only renders the drift_detected
// toggle for these kinds.
export const RESOURCES_WITH_DRIFT_DETECTED: ReadonlySet<ResourceKind> = new Set<ResourceKind>([
  'components',
  'sandboxes',
])

const SUB_OP_LABELS: Record<string, string> = {
  // shared lifecycle ops
  provision: 'Provision',
  deprovision: 'Deprovision',
  reprovision: 'Reprovision',
  // components
  deploy: 'Deploy',
  teardown: 'Teardown',
  // install_configurations
  inputs: 'Input updates',
  secrets: 'Secret syncs',
  // runners
  inactive: 'Inactive',
  // actions
  run: 'Run',
}

export const labelForSubOp = (op: string): string =>
  SUB_OP_LABELS[op] ??
  op.replace(/_/g, ' ').replace(/^./, (c) => c.toUpperCase())
